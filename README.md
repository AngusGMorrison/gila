# Gila
A high-performance, Vim-like text editor with minimal dependencies inspired by antirez's
[kilo](https://github.com/antirez/kilo).

Gila is currently a work in progress, but feel free to try it out. Simply clone the repo, run `make
build` and run the exported binary with the path to a text file as its first argument. Sample files
are provided under `testdata`.

## Progress
- [x] Enable terminal raw mode
- [x] Display welcome screen
- [x] Read and transliterate special keypresses, e.g. arrow keys
- [x] Cursor control
- [x] Load arbitrary text files
- [x] Vertical scrolling
- [ ] Horizontal scrolling
- [ ] Word wrap
- [ ] Text editing
- [ ] Search
- [ ] Syntax highlighting
- [ ] Mode-based editing (Vim style)
- [ ] Test suite
