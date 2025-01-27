// Copyright © 2016-2019 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/shenwei356/breader"
	"github.com/shenwei356/util/byteutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// tab2faCmd represents the tab2fx command
var tab2faCmd = &cobra.Command{
	Use:   "tab2fx",
	Short: "convert tabular format to FASTA/Q format",
	Long: `convert tabular format (first two/three columns) to FASTA/Q format

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		lineWidth := config.LineWidth
		outFile := config.OutFile
		runtime.GOMAXPROCS(config.Threads)

		files := getFileList(args, true)

		commentPrefixes := getFlagStringSlice(cmd, "comment-line-prefix")

		outfh, err := xopen.Wopen(outFile)
		checkError(err)
		defer outfh.Close()

		type Slice []string
		fn := func(line string) (interface{}, bool, error) {
			line = strings.TrimRight(line, "\r\n")

			if line == "" {
				return "", false, nil
			}
			// check comment line
			isCommentLine := false
			for _, p := range commentPrefixes {
				if strings.HasPrefix(line, p) {
					isCommentLine = true
					break
				}
			}
			if isCommentLine {
				return "", false, nil
			}

			items := strings.Split(line, "\t")
			if len(items) < 2 {
				return Slice(items), false, fmt.Errorf("at least two columns needed: %s", line)
			}
			if len(items) > 2 {
				return Slice(items[0:3]), true, nil
			}
			return Slice(items[0:2]), true, nil
		}

		for _, file := range files {
			reader, err := breader.NewBufferedReader(file, config.Threads, 10, fn)
			checkError(err)
			var text []byte
			var b *bytes.Buffer
			isFastq := false
			for chunk := range reader.Ch {
				if chunk.Err != nil {
					checkError(chunk.Err)
					break
				}
				for _, data := range chunk.Data {
					items := data.(Slice)
					if len(items) == 3 && (len(items[2]) > 0 || isFastq) { // fastq
						isFastq = true
						outfh.WriteString(fmt.Sprintf("@%s\n", items[0]))

						// 	outfh.Write(byteutil.WrapByteSlice([]byte(items[1]), lineWidth))

						// if bufferedByteSliceWrapper == nil {
						// 	bufferedByteSliceWrapper = byteutil.NewBufferedByteSliceWrapper2(1, len(items[1]), lineWidth)
						// }
						// text, b = bufferedByteSliceWrapper.Wrap([]byte(items[1]), lineWidth)
						// outfh.Write(text)
						// outfh.Flush()
						// bufferedByteSliceWrapper.Recycle(b)

						outfh.WriteString(items[1]) // seq

						outfh.WriteString("\n+\n")

						// outfh.Write(byteutil.WrapByteSlice([]byte(items[2]), lineWidth))

						// text, b = bufferedByteSliceWrapper.Wrap([]byte(items[2]), lineWidth)
						// outfh.Write(text)
						// outfh.Flush()
						// bufferedByteSliceWrapper.Recycle(b)

						outfh.WriteString(items[2]) // qual

						outfh.WriteString("\n")
					} else {
						outfh.WriteString(fmt.Sprintf(">%s\n", items[0]))

						// outfh.Write(byteutil.WrapByteSlice([]byte(items[1]), lineWidth))
						if bufferedByteSliceWrapper == nil {
							bufferedByteSliceWrapper = byteutil.NewBufferedByteSliceWrapper2(1, len(items[1]), lineWidth)
						}
						text, b = bufferedByteSliceWrapper.Wrap([]byte(items[1]), lineWidth)
						outfh.Write(text)
						outfh.Flush()
						bufferedByteSliceWrapper.Recycle(b)

						outfh.WriteString("\n")
					}
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(tab2faCmd)
	tab2faCmd.Flags().StringSliceP("comment-line-prefix", "p", []string{"#", "//"}, "comment line prefix")
}
