# tracedb [![GoDoc](https://godoc.org/github.com/unit-io/tracedb?status.svg)](https://godoc.org/github.com/unit-io/tracedb) [![Go Report Card](https://goreportcard.com/badge/github.com/unit-io/tracedb)](https://goreportcard.com/report/github.com/unit-io/tracedb) [![Coverage Status](https://coveralls.io/repos/github/unit-io/tracedb/badge.svg?branch=master)](https://coveralls.io/github/unit-io/tracedb?branch=master)

<p align="left">
  <img src="tracedb.png" width="70" alt="Trace" title="tracedb: Blazing fast timeseries database fro IoT and real-time gaming application"> 
</p>

# tracedb: Blazing fast timeseries database for IoT and real-time gaming applications

tracedb is blazing fast timeseries database for IoT, realtime messaging  applications. Access tracedb with pubsub over tcp or websocket using [trace](https://github.com/unit-io/trace) application.

Tracedb can be used for online gaming and mobile apps as it satisfy the requirements for low latency and binary messaging. Tracedb is perfect timeseries data store for applications such as internet of things and internet connected devices.

# Key characteristics
- 100% Go.
- Optimized for fast lookups and bulk inserts.
- Can store larger-than-memory data sets.
- Low memory usage.
- All DB methods are safe for concurrent use by multiple goroutines.

# Planned
- Memory buffer optimization to achive hyper scale writes & reads. Memory dump to archive files to offload buffers and free memory
- End to end lifecycle management of message entry, to better inform client about entry state (using system topic) such as entry is in commited state, error state or expired state etc..
- Add system topics (read only topics) to notify clients. For example topic -> "system/errors" to send realtime detailed error messages to client or notify if error has recoverd
- Documentation - document the technical atchitecture, design principals and advanced usage guides such as optimum configuration guideline to acive maximum throughput for hyper scale writes/reads operations (without bloting memory buffers).

## Quick Start
To build tracedb from source code use go get command.

> go get -u github.com/unit-io/tracedb

## Usage

### Opening a database

To open or create a new database, use the tracedb.Open() function:


```
package main

import (
	"log"

	"github.com/unit-io/tracedb"
)

func main() {
    db, err := tracedb.Open("tracedb.example", nil)
    if err != nil {
        log.Fatal(err)
        return
    }	
    defer db.Close()
}
```

### Writing to a database
Use the DB.PutEntry or DB.Batch() function to store messages to topic or delete a message from topic with DB.DeleteEntry function. Batch operation is non-blocking so client program can decide to wait for completed signal and further execute any additional tasks. Batch operation speeds up bulk record insertion into tracedb. Reading data is blazing fast if batch operation is used for bulk insertion and then reading records within short span of time while db is still open. See benchmark examples and run it locally to see performance of runnig batches concurrently.

Batch operations are optimized to high scale writes so if your client application perform few hundred writes per 5 milliseconds then choose some other DBs as tracedb may not be right solutions.

```
    err := db.Batch(func(b *tracedb.Batch, completed <-chan struct{}) error {
		b.Put([]byte("unit8.b.b1"), []byte("msg.b.b11.1"))
		b.Put([]byte("unit8.b.b11"), []byte("msg.b.b11.2"))
		b.Put([]byte("unit8.b.*"), []byte("msg.b.*.1"))
		b.Put([]byte("unit8.b.*"), []byte("msg.b.*.2"))
		err := b.Write()
		if !nonblocking {
			go func() {
				<-completed // it signals batch has completed and fully committed to log
				log.Printf("batch completed")
				print([]byte("unit8.b.b1?last=30m"), db)
				print([]byte("unit8.b.b11?last=30m"), db)
			}()
		}
		return err
    })

```

Deleting message.
Deleting a message in tracedb is rare and it require additional steps to delete message from given topic. Generate a unique message ID using DB.NewID() and use this unique message ID while putting message to the tracedb using DB.PutEntry(). To delete message provide message ID to the DB.DeleteEntry() fucntion.

```

	messageId := db.NewID()
	err := db.PutEntry(&tracedb.Entry{
		ID:       messageId,
		Topic:    []byte("unit8.b.b1"),
		Payload:  []byte("msg.b.b1.2"),
		Contract: 3376684800,
	})
	
	err := db.DeleteEntry(&tracedb.Entry{
		ID:       messageId,
		Topic:    []byte("unit8.b.b1"),
		Contract: 3376684800,
	})

```

Writing to wildcard topics.
Tracedb supports wrting to wildcard topics. Use "`*`" in the topic to write to wildcard topic or use "`...`" at the end of topic to write to all sub-topics. Writing to following wildcard topics are also supported, "`*`" or "`...`"

```
	err := db.Batch(func(b *tracedb.Batch, completed <-chan struct{}) error {
		b.Put([]byte("unit8.*.b11"), []byte("msg.*.b11.1"))
		b.Put([]byte("unit8.b.*"), []byte("msg.b.*.1"))
		b.Put([]byte("unit8..."), []byte("msg...1"))
		b.Put([]byte("*"), []byte("msg.*.1"))
		b.Put([]byte("..."), []byte("msg...1"))
		err := b.Write()
		return err
    })

```

Specify ttl to expires keys. 

```
err := db.Batch(func(b *tracedb.Batch, completed <-chan struct{}) error {
		b.Put([]byte("unit8.b.b1?ttl=3m"), []byte("msg.b.b1.1"))
		b.Put([]byte("unit8.b.b11?ttl=3m"), []byte("msg.b.b11.1"))
		b.Put([]byte("unit8.b.b111?ttl=3m"), []byte("msg.b.b111.1"))
		err := b.Write()
		return err
	})
```

To encrypt messages use batch options and set message encryption. Note, encryption can also be set on entire database using DB.Open() and provide encryption in the option parameter.

```
err := db.Batch(func(b *tracedb.Batch, completed <-chan struct{}) error {
		opts := tracedb.DefaultBatchOptions
		opts.Encryption = true
		b.SetOptions(opts)
		b.Put([]byte("unit8.b.b1?ttl=3m"), []byte("msg.b.b1.1"))
		b.Put([]byte("unit8.b.b11?ttl=3m"), []byte("msg.b.b11.1"))
		b.Put([]byte("unit8.b.b111?ttl=3m"), []byte("msg.b.b111.1"))
		err := b.Write()
		return err
	})
```

Use the BatchGroup.Add() function to group batches and run concurrently without causing write conflict. Use the BatchGroup.Run to run group of batches concurrently:

```
    g := db.NewBatchGroup()
	g.Add(func(b *tracedb.Batch, completed <-chan struct{}) error {
		b.Put([]byte("unit8.b.b1?ttl=2m"), []byte("msg.b.b1.1"))
		b.Put([]byte("unit8.c.c1?ttl=1m"), []byte("msg.c.c1.1"))
		b.Put([]byte("unit8.b.b1?ttl=3m"), []byte("msg.b.b1.2"))
		b.Write()
		go func() {
			<-completed // it signals batch group completion
			log.Printf("batch group completed")
		}()
		return nil
	})

	g.Add(func(b *tracedb.Batch, completed <-chan struct{}) error {
		b.Put([]byte("unit8.b.b11"), []byte("msg.b.b11.1"))
		b.Put([]byte("unit8.b.b11"), []byte("msg.b.b11.2"))
		b.Put([]byte("unit8.b.b1"), []byte("msg.b.b1.3"))
		b.Put([]byte("unit8.c.c11"), []byte("msg.c.c11.1"))
		b.Write()
		go func() {
			<-completed // it signals batch group completion
			log.Printf("batch group completed")
		}()
		return nil
	})

	g.Add(func(b *tracedb.Batch, completed <-chan struct{}) error {
		b.Put([]byte("unit8.b.b111"), []byte("msg.b.b111.1"))
		b.Put([]byte("unit8.b.b111"), []byte("msg.b.b111.2"))
		b.Put([]byte("unit8.b.b11"), []byte("msg.b.b11.3"))
		b.Put([]byte("unit8.c.c111"), []byte("msg.c.c111"))
		b.Write()
		go func() {
			<-completed // it signals batch group completion
			log.Printf("batch group completed")
		}()
		return nil
	})

	err = g.Run()

```

```
    func(retry int) {
		i := 1
		for j := range time.Tick(60 * time.Second) {
			print([]byte("unit8.b.b1?last=2m"), db)
			print([]byte("unit8.b.b11?last=2m"), db)
			print([]byte("unit8?last=2m"), db)
			print([]byte("unit9?last=2m"), db)
			if i >= retry {
				return
			}
			i++
		}
	}(4)
```

### Iterating over items
Use the DB.Items() function which returns a new instance of ItemIterator. 
Specify topic to retrives values and use last parameter to specify duration or specify number of recent messages to retreive from the topic. for example, "last=1h" retrieves messsages from tracedb stored in last 1 hour, or "last=100" to retreives last 100 messages from the tracedb:

```
func print(topic []byte, db *tracedb.DB) {
	// topic -> "unit8.b.b1?last=10m"
	it, err := db.Items(&tracedb.Query{Topic: topic})
	if err != nil {
		log.Fatal(err)
		return
	}
	for it.First(); it.Valid(); it.Next() {
		err := it.Error()
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Printf("%s %s", it.Item().Topic(), it.Item().Value())
	}
}
```

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2019 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/tracedb/blob/master/LICENSE).
