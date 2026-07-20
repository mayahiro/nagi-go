// Package text provides Unicode grapheme and terminal cell text primitives for
// Nagi.
//
// Segmentation follows Unicode Standard Annex #29 extended grapheme clusters
// using the committed Unicode data version. Width is an explicit terminal
// policy. Unicode normalization is not applied, and invalid UTF-8 runs are
// replaced before text operations.
package text
