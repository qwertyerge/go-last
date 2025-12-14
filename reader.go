package main

import (
	"io"
	"os"
)

const (
	// BUFFER_SIZE is the size of the read buffer (16KB)
	BUFFER_SIZE = 16 * 1024
)

// WtmpReader reads wtmp/btmp files from end to beginning
type WtmpReader struct {
	file       *os.File
	fileSize   int64
	position   int64
	buffer     []byte
	bufferPos  int
	bufferSize int
	recordSize int
	partial    []byte // Partial record from previous buffer
}

// NewWtmpReader creates a new WtmpReader
func NewWtmpReader(filename string) (*WtmpReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	recordSize := (&Utmpx{}).Size()
	
	return &WtmpReader{
		file:       file,
		fileSize:   stat.Size(),
		position:   stat.Size(),
		buffer:     make([]byte, BUFFER_SIZE),
		recordSize: recordSize,
	}, nil
}

// Close closes the underlying file
func (r *WtmpReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// ReadRecord reads the next record (backwards from end)
func (r *WtmpReader) ReadRecord() (*Utmpx, error) {
	for {
		// Check if we need to read a new buffer
		if r.bufferPos <= 0 {
			if err := r.readBuffer(); err != nil {
				if err == io.EOF && len(r.partial) > 0 {
					// Return partial record if it's complete
					if len(r.partial) == r.recordSize {
						record := &Utmpx{}
						if err := record.FromBytes(r.partial); err == nil {
							r.partial = nil
							return record, nil
						}
					}
					return nil, io.EOF
				}
				return nil, err
			}
		}

		// Calculate how many bytes we can read from current position
		bytesToRead := r.bufferPos
		
		// Align to record boundaries
		if bytesToRead >= r.recordSize {
			// Read complete record(s) from buffer
			// Get one record
			recordStart := r.bufferPos - r.recordSize
			recordData := make([]byte, r.recordSize)
			copy(recordData, r.buffer[recordStart:r.bufferPos])
			
			// If we have a partial from before, prepend it
			if len(r.partial) > 0 {
				combined := append(recordData, r.partial...)
				r.partial = nil
				
				// Now we might have more than one record
				if len(combined) >= r.recordSize {
					// Take the last complete record
					numRecords := len(combined) / r.recordSize
					lastRecordStart := (numRecords - 1) * r.recordSize
					recordData = combined[lastRecordStart : lastRecordStart+r.recordSize]
					
					// Keep the rest as partial
					if lastRecordStart > 0 {
						r.partial = combined[:lastRecordStart]
					}
				}
			}
			
			r.bufferPos = recordStart
			
			record := &Utmpx{}
			if err := record.FromBytes(recordData); err != nil {
				return nil, err
			}
			return record, nil
			
		} else {
			// We have a partial record, save it and read next buffer
			partial := make([]byte, bytesToRead)
			copy(partial, r.buffer[:bytesToRead])
			r.partial = append(partial, r.partial...)
			r.bufferPos = 0
		}
	}
}

// readBuffer reads the next buffer from file (backwards)
func (r *WtmpReader) readBuffer() error {
	if r.position <= 0 {
		return io.EOF
	}

	// Calculate how much to read
	toRead := int64(BUFFER_SIZE)
	if r.position < toRead {
		toRead = r.position
	}

	// Seek to position
	r.position -= toRead
	_, err := r.file.Seek(r.position, io.SeekStart)
	if err != nil {
		return err
	}

	// Read into buffer
	n, err := io.ReadFull(r.file, r.buffer[:toRead])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return err
	}

	if n == 0 {
		return io.EOF
	}

	r.bufferSize = n
	r.bufferPos = n

	return nil
}

// HasMore checks if there are more records to read
func (r *WtmpReader) HasMore() bool {
	return r.position > 0 || r.bufferPos > 0 || len(r.partial) > 0
}

// ReadAllRecords reads all records from the file (backwards)
func ReadAllRecords(filename string) ([]*Utmpx, error) {
	reader, err := NewWtmpReader(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var records []*Utmpx
	for {
		record, err := reader.ReadRecord()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if record != nil {
			records = append(records, record)
		}
	}

	return records, nil
}
