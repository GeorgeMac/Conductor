# Conductor

## A tiny Golang service / program framework

Orchestrate asynchronous functions / closeable services with a Conductor

``` Go
    cond := conductor.NewConductor(func(err error) { log.Error(err) })

    fi, _ ioutil.ReadFile(“path/to/file”)

    db, _ := db.Connect(“blah...”)

    cond.RegisterClosers(fi, db)

    cond.Go(func(){
        // Do something that might take a while
    })

    cond.Go(func(){
        // Do something else that might take a while
    })

    cond.FinishAsap()
    //or
    cond.FinishOnSignals(syscall.SIGTERM, syscall.SIGINT)
```
