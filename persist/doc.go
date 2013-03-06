// Package persist implements the record wizard's persistence layer.
//
// Currently persist reads and writes a trio of flat JSON files:
//
//     questions.json
//     forms.json
//     records.json
//
// Package persist is goroutine-safe; it serializes all requests through an
// internal channel.
package persist
