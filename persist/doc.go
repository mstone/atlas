// Package persist implements the review wizard's persistence layer.
//
// Currently persist reads and writes a trio of flat JSON files:
//
//     questions.json
//     profiles.json
//     reviews.json
//
// Package persist is goroutine-safe; it serializes all requests through an
// internal channel.
package persist
