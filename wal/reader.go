/*
 * Copyright 2020 Saffat Technologies, Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wal

import (
	"encoding/binary"
	"errors"

	"github.com/unit-io/bpool"
	"github.com/unit-io/unitdb/uid"
)

// Reader reads logs from WAL.
// Reader reader is a simple iterator over log data.
type Reader struct {
	Id         uid.LID
	offset     int64
	entryCount uint32
	buffer     *bpool.Buffer

	wal *WAL
}

// NewReader returns new log reader to read the logs from the WAL.
func (wal *WAL) NewReader() (*Reader, error) {
	if err := wal.ok(); err != nil {
		return &Reader{wal: wal}, err
	}
	r := &Reader{
		Id:  uid.NewLID(),
		wal: wal,
	}

	return r, nil
}

// Iterator iterates the pending logs from the WAL.
func (r *Reader) Iterator(f func(timeID int64) (bool, error)) (err error) {
	r.wal.mu.RLock()
	r.buffer = r.wal.bufPool.Get()
	defer func() {
		r.wal.recoveredTimeIDs = r.wal.recoveredTimeIDs[:0]
		r.wal.bufPool.Put(r.buffer)
		r.wal.mu.RUnlock()
	}()

	for _, timeID := range r.wal.recoveredTimeIDs {
		r.offset = 0
		r.buffer.Reset()
		info := r.wal.logStore.read(timeID, r.buffer)
		r.entryCount = info.count
		if stop, err := f(timeID); stop || err != nil {
			return err
		}
	}

	return nil
}

// Count returns entry count for the current interation.
func (r *Reader) Count() uint32 {
	return r.entryCount
}

// Next returns next record from the iterator or false if iteration is done.
func (r *Reader) Next() ([]byte, bool, error) {
	if r.entryCount == 0 {
		return nil, false, nil
	}
	r.entryCount--
	scratch, _ := r.buffer.Slice(r.offset, r.offset+4)
	dataLen := binary.LittleEndian.Uint32(scratch)
	data, err := r.buffer.Slice(r.offset+4, r.offset+int64(dataLen))
	if err != nil {
		return nil, false, errors.New("error reading log")
	}
	r.offset += int64(dataLen)
	return data, true, nil
}
