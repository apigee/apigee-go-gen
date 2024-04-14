//  Copyright 2024 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Inspired by Eli Bendersky [https://eli.thegreenplace.net]

package utils

import (
	"bytes"
	"io"
	"log"
	"os"
)

type StdoutCapture struct {
	realStdout *os.File
	fakeStdout *os.File
	out        chan []byte
}

func NewStdoutCapture() (*StdoutCapture, error) {

	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	stdout := os.Stdout
	os.Stdout = writer
	outCh := make(chan []byte)

	go func() {
		var b bytes.Buffer
		if _, err := io.Copy(&b, reader); err != nil {
			log.Println(err)
		}
		outCh <- b.Bytes()
	}()

	return &StdoutCapture{
		realStdout: stdout,
		fakeStdout: reader,
		out:        outCh,
	}, nil
}

func (sf *StdoutCapture) Read() ([]byte, error) {
	err := os.Stdout.Close()
	if err != nil {
		return nil, err
	}
	out := <-sf.out
	return out, nil
}

func (sf *StdoutCapture) Restore() {
	os.Stdout = sf.realStdout
}
