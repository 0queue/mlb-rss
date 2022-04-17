# THOUGHTS.md

Quick retrospective on the development process. The implementation
took about two days, or one and half rather, with this reflection
being written the following day.

## Go

This is my first Go program, not counting some small experiments
with Caddy.

Overall I confirmed my distaste for the Go style of programming,
and design choices made in the language itself, compared to the
other languages in my Github repos.  It's nice to have a little
experience to back it up though.  

That said, I do appreciate it as a step up from Python, which is
probably what I would consider next for this project if it weren't
for Rust. Static binaries and a static type system are a big step
up in my opinion.

## Helix

Another unusual choice for me in this project was using [Helix]
as the editor.  I chose Helix because it was very easy to set up
with a language server, whereas my Neovim configuration still has
a bit of work needed there.

I am very impressed by how usable it is already.  I did fumble a
lot the entire way through, because who has time to read documentation,
but it got the job done.  In no particular order, I enjoyed:

  - Language server usage
  - "Menus" with short descriptions of the commands
  - Lots of nice themes that are easy to switch between
  - Pretty easy to get started with vim knowledge, even
    though the command grammar is different
    
Things I missed:

  - Integrated terminal (or even something to run a command and see the
    output without editing the buffer)
  - Remembering edit location (?)
  
Things I should've looked up sooner:

  -  `mi"c` as a replacement for `ci"`

[Helix]: https://github.com/helix-editor/helix
