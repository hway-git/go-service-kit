package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrMessage = errors.New("邮件参数无效，请检查发件人、收件人、抄送、主题和正文类型")

func singleLine(value string) bool { return !strings.ContainsAny(value, "\r\n\x00") }

type Message struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Cc          string `json:"cc"` // 抄送地址，可为空。
	Subject     string `json:"subject"`
	ContentType string `json:"content_type"` // 缺省为text/html，也支持text/plain。
	Body        string `json:"body"`
}

type Response struct {
	Code    int    `json:"code"` // 0表示成功，-1表示失败。
	Message string `json:"message"`
}

// HandleJSON 保留邮件消息与响应字段，调用方负责订阅及回传处理结果。
func (c *Client) HandleJSON(ctx context.Context, data []byte) *Response {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var message *Message
	if decoder.Decode(&message) != nil || message == nil || decoder.Decode(new(any)) != io.EOF {
		return &Response{Code: -1, Message: ErrMessage.Error()}
	}
	if err := c.Send(ctx, message); err != nil {
		return &Response{Code: -1, Message: err.Error()}
	}
	return &Response{Code: 0, Message: "success"}
}

func addresses(value string, optional bool) ([]*mail.Address, error) {
	if optional && value == "" {
		return nil, nil
	}
	if !singleLine(value) {
		return nil, ErrMessage
	}
	parsed, err := mail.ParseAddressList(value)
	if err != nil || len(parsed) == 0 {
		return nil, ErrMessage
	}
	return parsed, nil
}

func addressHeader(addresses []*mail.Address) string {
	values := make([]string, len(addresses))
	for index, address := range addresses {
		values[index] = address.String()
	}
	return strings.Join(values, ", ")
}

func (c *Client) compose(message *Message) (string, []string, []byte, error) {
	if message == nil || !singleLine(message.From) || !singleLine(message.Subject) || !utf8.ValidString(message.Subject) || !utf8.ValidString(message.Body) {
		return "", nil, nil, ErrMessage
	}
	from := message.From
	if from == "" {
		from = c.config.EmailFrom
	}
	sender, err := mail.ParseAddress(from)
	if err != nil {
		return "", nil, nil, ErrMessage
	}
	to, err := addresses(message.To, false)
	if err != nil {
		return "", nil, nil, err
	}
	cc, err := addresses(message.Cc, true)
	if err != nil {
		return "", nil, nil, err
	}
	contentType := message.ContentType
	if contentType == "" {
		contentType = "text/html"
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || (mediaType != "text/html" && mediaType != "text/plain") || len(params) > 1 || (len(params) == 1 && !strings.EqualFold(params["charset"], "utf-8")) {
		return "", nil, nil, ErrMessage
	}
	var body bytes.Buffer
	fmt.Fprintf(&body, "From: %s\r\nTo: %s\r\n", sender.String(), addressHeader(to))
	if len(cc) != 0 {
		fmt.Fprintf(&body, "Cc: %s\r\n", addressHeader(cc))
	}
	fmt.Fprintf(&body, "Subject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", mime.QEncoding.Encode("UTF-8", message.Subject), time.Now().Format(time.RFC1123Z), mediaType)
	writer := quotedprintable.NewWriter(&body)
	if _, err := writer.Write([]byte(message.Body)); err != nil {
		return "", nil, nil, ErrMessage
	}
	if writer.Close() != nil {
		return "", nil, nil, ErrMessage
	}
	var recipients []string
	seen := make(map[string]bool)
	for _, address := range append(to, cc...) {
		if !seen[address.Address] {
			recipients = append(recipients, address.Address)
			seen[address.Address] = true
		}
	}
	return sender.Address, recipients, body.Bytes(), nil
}
