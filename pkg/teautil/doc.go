// Package teautil is a collection of helpful [tea] utilities.
//
// The bubbletea command and messaging system is completely devoid of any sense
// of opinionated message routing. Message simply “are”, and your top-level
// model just has to know where they are supposed to go.  The [Router] object
// solves this dilemma by drawing a line in the sand: any messages that result
// from a [tea.Cmd] returned by a child model should be routed back to that
// child model.
package teautil
