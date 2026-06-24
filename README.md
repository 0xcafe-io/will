will - a small Go library to use `defer` guilt-free.

## In a nutshell:

```go
defer will.CaptureErr(file.Close, &err) // will capture Close() error and merge it into err
defer will.LogErr(tx.Rollback)          // will log the Rollback() error if it fails
defer will.IgnoreErr(resp.Body.Close)   // ignore the error explicitly
defer will.RecoverTo(&err)              // will calm down the panics and recover them into err
```

## Use Cases

### Capture errors from deferred calls

Use `defer will.CaptureErr()` to capture errors from functions like `file.Close()`/`rows.Close()`/`buf.Flush()` and
merge them into returned error of the caller function.  
(error must be named to allow referencing)

```go
func writeEventsBatch(w io.Writer, events []string) (err error) {
    buf := bufio.NewWriter(w)
    // buf.Flush() error is critical, and must be reported
    defer will.CaptureErr(buf.Flush, &err)
    
    for _, event := range events {
        if _, err := buf.WriteString(event + "\n"); err != nil {
            return fmt.Errorf("failed to write event: %w", err)
        }
    }
    
    return nil 
}
```

You can use `will.CaptureErr()` multiple times in a single function.   
It uses `errors.Join()` from standard library to merge errors, therefore `errors.Is()` its friends will work as
expected.


---

### Log errors from deferred calls
Sometimes errors from deferred calls should not be propagated to the caller.   
But you still want to be aware of them.
Use `defer will.LogErr()` to log such errors.
```go
func (db *sql.DB) GetActiveUserIDs(ctx context.Context) ([]uuid.UUID, error) {
    rows, err := db.Query("SELECT id FROM users WHERE active")
    if err != nil {
        return nil, fmt.Errorf("failed to query active user ids: %w", err)
    }
    defer will.LogErr(rows.Close) // if Close() fails, err will be logged via log.Println()
    
    // scanning rows
    
    return ids, nil
}
```

Another with mixed error handling:

```go
func copyFile(srcPath, dstPath string) (err error) {
    srcFile, err := os.Open(srcPath)
    if err != nil {
        return fmt.Errorf("failed to open src file: %w", err)
    }
    defer will.LogErr(srcFile.Close) // potential error #1, not critical to the caller
    
    dstFile, err := os.Create(dstPath)
    if err != nil {
        return fmt.Errorf("failed to create dest file: %w", err)
    }
    defer will.CaptureErr(dstFile.Close, &err) // potential error #2, critical to the caller (might not hit the disk)
    
    _, err = io.Copy(dstFile, srcFile)
    
    return err
}


```

---

### Custom error handling

Use `will.HandleErr()` to handle errors from deferred calls with a custom pre-configured handler.  
(E.g custom logging, metrics, etc.)

```go
// main.go
func main() {
    will.SetErrHandler(func(err error) {
        slog.Error(err.Error())
        sentry.CaptureException(err)
        metrics.ErrorsTotal.Inc()
    })
}

// worker.go
func processJob(job Job) {
    // ...
    defer will.HandleErr(job.Cleanup) // unexpected cleanup failures are reported to Sentry/metrics
    // ...
}
```

---

### Ignore explicitly
Don't give a fuck about errors from deferred calls?  
Make your intention clear to other developers and linters by using `defer will.IgnoreErr()`.

```go
func loadConfig(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open config file: %w", err)
    }
    defer will.IgnoreErr(f.Close) // rare and not actionable, just ignore
    
    var cfg Config
    if err := json.NewDecoder(f).Decode(&cfg); err != nil {
        return nil, fmt.Errorf("failed to decode config: %w", err)
    }
    
    return &cfg, nil
}

```

---

### Recover from Panic

Use `defer will.RecoverTo()` to recover from panics and capture them into the returned error.

```go
func runTask(ctx context.Context, task Task) (err error) {
    // Prevent a single task from crashing the entire worker pool.
    // If task.Run panics, it will be recovered and returned as err.
    defer will.RecoverTo(&err)
    
    return task.Run(ctx)
}
```

---

## Install

```bash
go get github.com/0xcafe-io/will
```

