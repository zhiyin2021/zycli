package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zhiyin2021/zycli/tools"
)

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*logWriter)(nil)

type CompressType string

const (
	CT_NONE CompressType = ""
	CT_GZ   CompressType = ".gz"
	CT_XZ   CompressType = ".xz"
)

type logWriter struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	filename string

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	maxSize int64

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	maxAge int

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	maxCount int

	// // LocalTime determines if the time used for formatting the timestamps in
	// // backup files is the computer's local time.  The default is to use UTC
	// // time.
	// localTime bool
	// layout string
	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	compressType CompressType
	// BackupTimeFormat string `json:"backuptimeformat" yaml:"backuptimeformat"`
	size int64
	file *os.File
	mu   sync.Mutex

	millCh     chan bool
	startMill  sync.Once
	millRuning int32
	dir        string

	ctime time.Time
	pos   int
}

var (
	// os_Stat exists so it can be mocked out by tests.
	osStat = os.Stat

	// megabyte is the conversion factor between MaxSize and bytes.  It is a
	// variable so tests can mock it out and not need to write megabytes of data
	// to disk.
	mbyte int64 = 1024 * 1024
)

type logWriterOption func(l *logWriter)

func NewSplit(fileName string, opts ...logWriterOption) *logWriter {
	if fileName == "" {
		fileName = filepath.Join(os.TempDir(), filepath.Base(os.Args[0])+".log")
	}
	l := &logWriter{
		filename:     fileName,
		maxSize:      100 * mbyte, // megabytes
		maxCount:     0,
		maxAge:       31,    // days
		compressType: CT_GZ, // disabled by default
		// layout:       "060102150405.000",
		millRuning: 0,
		dir:        filepath.Dir(fileName),
		pos:        0,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// 最大切割文件大小,单位:MB,默认:100MB
func OptMaxSize(maxSize int64) logWriterOption {
	return func(l *logWriter) {
		l.maxSize = maxSize * mbyte
	}
}

// 保留文件数量,默认:0
func OptMaxCount(maxCount int) logWriterOption {
	return func(l *logWriter) {
		l.maxCount = maxCount
	}
}

// 保留天数,默认:31
func OptMaxAge(maxAge int) logWriterOption {
	return func(l *logWriter) {
		l.maxAge = maxAge
	}
}

// 是否压缩,默认:是
func OptCompressType(compressType CompressType) logWriterOption {
	return func(l *logWriter) {
		l.compressType = compressType
	}
}

// 切割文件时间格式,默认:060102150405.000
// func OptLayout(layout string) logWriterOption {
// 	return func(l *logWriter) {
// 		l.layout = layout
// 	}
// }

// Write implements io.Writer.  If a write would cause the log file to be larger
// than MaxSize, the file is closed, renamed to include a timestamp of the
// current time, and a new log file is created using the original log file name.
// If the length of the write is greater than MaxSize, an error is returned.
func (l *logWriter) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	writeLen := int64(len(p))
	if writeLen > l.maxSize {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, l.maxSize,
		)
	}

	if l.file == nil {
		if err = l.openExistingOrNew(len(p)); err != nil {
			return 0, err
		}

	}

	if l.size+writeLen > l.maxSize {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}
	if l.ctime.Day() != time.Now().Day() {
		l.pos = 0
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}
	n, err = l.file.Write(p)
	l.size += int64(n)

	return n, err
}

// Close implements io.Closer, and closes the current logfile.
func (l *logWriter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

// close closes the file if it is open.
func (l *logWriter) close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

// Rotate causes logWriter to close the existing log file and immediately create a
// new one.  This is a helper function for applications that want to initiate
// rotations outside of the normal rotation rules, such as in response to
// SIGHUP.  After rotating, this initiates compression and removal of old log
// files according to the configuration.
func (l *logWriter) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotate()
}

// rotate closes the current file, moves it aside with a timestamp in the name,
// (if it exists), opens a new file with the original filename, and then runs
// post-rotation processing and removal.
func (l *logWriter) rotate() error {
	if err := l.close(); err != nil {
		return err
	}
	if err := l.openNew(); err != nil {
		fmt.Println("rotate.openNew", err)
		return err
	}
	l.mill()
	return nil
}

// openNew opens a new log file for writing, moving any old log file out of the
// way.  This methods assumes the file has already been closed.
func (l *logWriter) openNew() error {
	err := os.MkdirAll(l.dir, 0755)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	name := l.filename
	mode := os.FileMode(0600)
	info, err := osStat(name)
	if err == nil {
		// Copy the mode off the old logfile.
		mode = info.Mode()
		// move the existing file
		newname := l.backupName(name)
		if err := os.Rename(name, newname); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}

		// this is a no-op anywhere but linux
		if err := tools.Chown(name, info); err != nil {
			return err
		}
	}

	// we use truncate here because this should only get called when we've moved
	// the file ourselves. if someone else creates the file in the meantime,
	// just wipe out the contents.
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	l.file = f
	l.size = 0
	l.ctime = time.Now()
	return nil
}

// backupName creates a new filename from the given name, inserting a timestamp
// between the filename and the extension, using the local time if requested
// (otherwise UTC).
func (l *logWriter) backupName(name string) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]

	if l.ctime.Day() != time.Now().Day() {
		l.ctime = time.Now()
		l.pos = 0
	}
	timestamp := l.ctime.Format("20060102")
	for {
		l.pos++
		logPath := filepath.Join(dir, fmt.Sprintf("%s_%s.%d%s", prefix, timestamp, l.pos, ext))
		if _, err := osStat(logPath + string(l.compressType)); err != nil {
			return logPath
		}
	}
}

// openExistingOrNew opens the logfile if it exists and if the current write
// would not put it over MaxSize.  If there is no such file or the write would
// put it over the MaxSize, a new file is created.
func (l *logWriter) openExistingOrNew(writeLen int) error {
	l.mill()

	filename := l.filename
	info, err := osStat(filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	if info.Size()+int64(writeLen) >= l.maxSize {
		return l.rotate()
	}
	if info.ModTime().Day() != time.Now().Day() {
		return l.rotate()
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return l.openNew()
	}
	l.file = file
	l.size = info.Size()
	l.ctime = info.ModTime()
	return nil
}

// millRunOnce performs compression and removal of stale log files.
// Log files are compressed if enabled via configuration and old log
// files are removed, keeping at most l.MaxBackups files, as long as
// none of them are older than MaxAge.
func (l *logWriter) millRunOnce() error {
	// fmt.Printf("millRunOnce:%#v\n", l)
	if ok := atomic.CompareAndSwapInt32(&l.millRuning, 0, 1); ok {
		defer atomic.StoreInt32(&l.millRuning, 0)
		if l.maxCount == 0 && l.maxAge == 0 && l.compressType == CT_NONE {
			return nil
		}

		files, err := l.oldLogFiles()
		if err != nil {
			return err
		}

		var compress, remove []fs.FileInfo

		if l.maxCount > 0 && l.maxCount < len(files) {
			preserved := make(map[string]bool)
			var remaining []fs.FileInfo
			for _, f := range files {
				// Only count the uncompressed log file or the
				// compressed log file, not both.
				fn := f.Name()
				if ok := strings.HasSuffix(fn, string(l.compressType)); ok {
					fn = fn[:len(fn)-len(l.compressType)]
				}
				preserved[fn] = true

				if len(preserved) > l.maxCount {
					remove = append(remove, f)
				} else {
					remaining = append(remaining, f)
				}
			}
			files = remaining
		}
		if l.maxAge > 0 {
			diff := time.Duration(int64(24*time.Hour) * int64(l.maxAge))
			cutoff := time.Now().Add(-1 * diff)

			var remaining []fs.FileInfo
			for _, f := range files {
				if f.ModTime().Before(cutoff) {
					remove = append(remove, f)
				} else {
					remaining = append(remaining, f)
				}
			}
			files = remaining
		}

		if l.compressType != CT_NONE {
			for _, f := range files {
				if !strings.HasSuffix(f.Name(), string(l.compressType)) {
					compress = append(compress, f)
				}
			}
		}

		for _, f := range remove {
			errRemove := os.Remove(filepath.Join(l.dir, f.Name()))
			if err == nil && errRemove != nil {
				err = errRemove
			}
		}
		for _, f := range compress {
			fn := filepath.Join(l.dir, f.Name())
			errCompress := l.compress(fn, fn+string(l.compressType))
			if err == nil && errCompress != nil {
				err = errCompress
			}
		}
		return err
	}
	return nil
}

// millRun runs in a goroutine to manage post-rotation compression and removal
// of old log files.
func (l *logWriter) millRun() {
	for range l.millCh {
		// what am I going to do, log this?
		_ = l.millRunOnce()
	}
}

// mill performs post-rotation compression and removal of stale log files,
// starting the mill goroutine if necessary.
func (l *logWriter) mill() {
	l.startMill.Do(func() {
		l.millCh = make(chan bool, 3)
		go l.millRun()
	})
	select {
	case l.millCh <- true:
	default:
	}
}

// oldLogFiles returns the list of backup log files stored in the same
// directory as the current log file, sorted by ModTime
func (l *logWriter) oldLogFiles() ([]fs.FileInfo, error) {
	files, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	logFiles := []fs.FileInfo{}

	prefix, ext := l.prefixAndExt()

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fi, _ := f.Info()
		// if t, err := l.timeFromName(f.Name(), prefix, ext); err == nil {
		// 	logFiles = append(logFiles, logInfo{t, fi})
		// 	continue
		// }
		fname := fi.Name()
		if strings.HasPrefix(fname, prefix) && (strings.HasSuffix(fname, ext) || strings.HasSuffix(fname, ext+string(l.compressType))) {
			logFiles = append(logFiles, fi)
			continue
		}
		// error parsing means that the suffix at the end was not generated
		// by lumberjack, and therefore it's not a backup file.
	}

	sort.Sort(byFormatTime(logFiles))

	return logFiles, nil
}

// prefixAndExt returns the filename part and extension part from the logWriter's
// filename.
func (l *logWriter) prefixAndExt() (prefix, ext string) {
	filename := filepath.Base(l.filename)
	ext = filepath.Ext(filename)
	prefix = filename[:len(filename)-len(ext)] + "_"
	return prefix, ext
}
func (l *logWriter) compress(src, dst string) (err error) {
	switch l.compressType {
	case CT_GZ:
		return gzcompress(src, dst)
	case CT_XZ:
		return xzcompress(src, dst)
	default:
		return nil
	}
}

// compressLogFile compresses the given log file, removing the
// uncompressed log file if successful.
func gzcompress(src, dst string) (err error) {
	return exec.Command("gzip", src).Start()
	// f, err := os.Open(src)
	// if err != nil {
	// 	return fmt.Errorf("failed to open log file: %v", err)
	// }
	// defer f.Close()

	// fi, err := osStat(src)
	// if err != nil {
	// 	return fmt.Errorf("failed to stat log file: %v", err)
	// }

	// if err := tools.Chown(dst, fi); err != nil {
	// 	return fmt.Errorf("failed to chown compressed log file: %v", err)
	// }

	// // If this file already exists, we presume it was created by
	// // a previous attempt to compress the log file.
	// gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	// if err != nil {
	// 	return fmt.Errorf("failed to open compressed log file: %v", err)
	// }
	// defer gzf.Close()

	// gz := gzip.NewWriter(gzf)

	// defer func() {
	// 	if err != nil {
	// 		os.Remove(dst)
	// 		err = fmt.Errorf("failed to compress log file: %v", err)
	// 	}
	// }()

	// if _, err := io.Copy(gz, f); err != nil {
	// 	return err
	// }
	// if err := gz.Close(); err != nil {
	// 	return err
	// }
	// if err := gzf.Close(); err != nil {
	// 	return err
	// }

	// if err := f.Close(); err != nil {
	// 	return err
	// }
	// if err := os.Remove(src); err != nil {
	// 	return err
	// }

	// return nil
}
func xzcompress(src, dst string) (err error) {
	return exec.Command("xz", "-z", src).Start()
	// f, err := os.Open(src)
	// if err != nil {
	// 	return fmt.Errorf("failed to open log file: %v", err)
	// }
	// defer func() {
	// 	f.Close()
	// 	if err == nil {
	// 		os.Remove(src)
	// 	}
	// }()
	// fi, err := osStat(src)
	// if err != nil {
	// 	return fmt.Errorf("failed to stat log file: %v", err)
	// }

	// if err := tools.Chown(dst, fi); err != nil {
	// 	return fmt.Errorf("failed to chown compressed log file: %v", err)
	// }

	// // If this file already exists, we presume it was created by
	// // a previous attempt to compress the log file.
	// xzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	// if err != nil {
	// 	return fmt.Errorf("failed to open compressed log file: %v", err)
	// }
	// defer func() {
	// 	xzf.Close()
	// 	if err != nil {
	// 		os.Remove(dst)
	// 		err = fmt.Errorf("failed to compress log file: %v", err)
	// 	}
	// }()
	// // const text = "The quick brown fox jumps over the lazy dog.\n"
	// // var buf bytes.Buffer
	// // // compress text
	// xzFile, err := xz.NewWriter(xzf)
	// if err != nil {
	// 	return err
	// }
	// defer xzFile.Close()
	// if _, err := io.Copy(xzFile, f); err != nil {
	// 	return err
	// }
	// return nil
}

// byFormatTime sorts by newest time formatted in the name.
type byFormatTime []os.FileInfo

func (b byFormatTime) Less(i, j int) bool {
	return b[i].ModTime().After(b[j].ModTime())
}

func (b byFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byFormatTime) Len() int {
	return len(b)
}
