# sass
Pure Go sass scanner, ast, and parser

Cross platform compiler for Sass

This project is currently in alpha, and contains no compiler. A scanner and parser are being developed to support a future compiler.

To help, check out [parser](https://github.com/wellington/sass/tree/master/parser). This project contains tests that iterate through sass-spec running the parser against example inputs. Errors detected by the parser are reported. However, you could also set the Parser mode to `Trace` and verify proper ast trees are being built from the input. As the parser matures, output can automatically be verified by the example outputs in these directories.
