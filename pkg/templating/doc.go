/*
Package templating provides a high-performance, filesystem-based Go template engine.
It is designed for generating dynamic, complex, and plausible-looking web content
with minimal server overhead.

It includes a rich library of custom template functions for generating
everything from structured data (JSON, forms) and styled elements (CSS, SVGs) to
thematic text via an integrated, SQLite-backed Markov generator. The engine is
fully configurable with safety limits to prevent abuse and supports hot-reloading
of templates from the filesystem, enabling easy updates post-deployment.

For a complete list of template functions and usage examples, see the README.md file.
*/
package templating
