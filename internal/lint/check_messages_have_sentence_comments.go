// Copyright (c) 2020 Uber Technologies, Inc.
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

package lint

import (
	"strings"

	"github.com/emicklei/proto"

	"github.com/uber/prototool/internal/text"
)

var messagesHaveSentenceCommentsLinter = NewLinter(
	"MESSAGES_HAVE_SENTENCE_COMMENTS",
	`Verifies that all non-extended messages types have a comment that contains at least one complete sentence.`,
	checkMessagesHaveSentenceComments,
)

func checkMessagesHaveSentenceComments(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&messagesHaveSentenceCommentsVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type messagesHaveSentenceCommentsVisitor struct {
	baseAddVisitor
	messageNameToMessage map[string]*proto.Message
	nestedMessageNames   []string
}

func (v *messagesHaveSentenceCommentsVisitor) OnStart(*FileDescriptor) error {
	v.messageNameToMessage = nil
	v.nestedMessageNames = nil
	return nil
}

func (v *messagesHaveSentenceCommentsVisitor) VisitMessage(message *proto.Message) {
	v.nestedMessageNames = append(v.nestedMessageNames, message.Name)
	for _, child := range message.Elements {
		child.Accept(v)
	}
	v.nestedMessageNames = v.nestedMessageNames[:len(v.nestedMessageNames)-1]

	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	if len(v.nestedMessageNames) > 0 {
		v.messageNameToMessage[strings.Join(v.nestedMessageNames, ".")+"."+message.Name] = message
	} else {
		v.messageNameToMessage[message.Name] = message
	}
}

func (v *messagesHaveSentenceCommentsVisitor) VisitService(service *proto.Service) {
	for _, child := range service.Elements {
		child.Accept(v)
	}
}

func (v *messagesHaveSentenceCommentsVisitor) Finally() error {
	if v.messageNameToMessage == nil {
		v.messageNameToMessage = make(map[string]*proto.Message)
	}
	for _, message := range v.messageNameToMessage {
		if !message.IsExtend {
			if !hasCompleteSentenceComment(message.Comment) {
				v.AddFailuref(message.Position, `Message %q needs a comment with a complete sentence that starts on the first line of the comment.`, message.Name)
			}
		}
	}
	return nil
}
