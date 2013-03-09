package linker

import (
	"bytes"
)

type LinkRenderer struct {
	Links []Link
}

func NewLinkRenderer() *LinkRenderer {
	return &LinkRenderer{}
}

type Link struct {
	Kind    int
	Href    string
	Title   string
	Alt     string
	Content string
}

const (
	AUTOLINK = iota
	LINK
	IMAGE
)

// block-level callbacks
func (self *LinkRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	return
}
func (self *LinkRenderer) BlockQuote(out *bytes.Buffer, text []byte) {
	return
}

func (self *LinkRenderer) BlockHtml(out *bytes.Buffer, text []byte) {
	return
}

func (self *LinkRenderer) Header(out *bytes.Buffer, text func() bool, level int) {
	text()
	return
}

func (self *LinkRenderer) HRule(out *bytes.Buffer) {
	return
}

func (self *LinkRenderer) List(out *bytes.Buffer, text func() bool, flags int) {
	text()
	return
}

func (self *LinkRenderer) ListItem(out *bytes.Buffer, text []byte, flags int) {
	return
}

func (self *LinkRenderer) Paragraph(out *bytes.Buffer, text func() bool) {
	text()
	return
}

func (self *LinkRenderer) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	return
}

func (self *LinkRenderer) TableRow(out *bytes.Buffer, text []byte) {
	return
}

func (self *LinkRenderer) TableCell(out *bytes.Buffer, text []byte, flags int) {
	return
}

// Span-level callbacks
func (self *LinkRenderer) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	self.Links = append(self.Links, Link{
		Kind: AUTOLINK,
		Href: string(link),
	})
}

func (self *LinkRenderer) CodeSpan(out *bytes.Buffer, text []byte) {
	return
}

func (self *LinkRenderer) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	return
}

func (self *LinkRenderer) Emphasis(out *bytes.Buffer, text []byte) {
	return
}

func (self *LinkRenderer) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	self.Links = append(self.Links, Link{
		Kind:  IMAGE,
		Href:  string(link),
		Title: string(title),
		Alt:   string(alt),
	})
}

func (self *LinkRenderer) LineBreak(out *bytes.Buffer) {
	return
}

func (self *LinkRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	self.Links = append(self.Links, Link{
		Kind:    LINK,
		Href:    string(link),
		Title:   string(title),
		Content: string(content),
	})
}

func (self *LinkRenderer) RawHtmlTag(out *bytes.Buffer, tag []byte) {
	return
}

func (self *LinkRenderer) TripleEmphasis(out *bytes.Buffer, text []byte) {
	return
}

func (self *LinkRenderer) StrikeThrough(out *bytes.Buffer, text []byte) {
	return
}

// Low-level callbacks
func (self *LinkRenderer) Entity(out *bytes.Buffer, entity []byte) {
	return
}

func (self *LinkRenderer) NormalText(out *bytes.Buffer, text []byte) {
	return
}

// Header and footer
func (self *LinkRenderer) DocumentHeader(out *bytes.Buffer) {
	return
}

func (self *LinkRenderer) DocumentFooter(out *bytes.Buffer) {
	return
}
