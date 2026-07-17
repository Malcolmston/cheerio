# Changelog

All notable changes to this project are documented here. The format is loosely
based on [Keep a Changelog](https://keepachangelog.com/), and the project aims
to follow semantic versioning.

## [0.2.0] - 2026-07-17

A large expansion toward feature parity with the Node.js cheerio package. Still
dependency-free and standard-library only.

### Added

- **Manipulation API**: `Append`, `Prepend`, `Before`, `After`, `AppendTo`,
  `PrependTo`, `InsertBefore`, `InsertAfter`, `ReplaceWith`, `Remove`, `Empty`,
  `Clone`, `Wrap`, `WrapAll`, `WrapInner`, `Unwrap`. Each content method also
  has `…Nodes` and `…Selection` variants (e.g. `AppendNodes`,
  `AppendSelection`). Multi-target inserts clone content; single-target inserts
  move it, matching jQuery semantics.
- **Setters**: `SetHtml`, `SetVal`, `SetProp` / `Prop`, `SetAttrs` (multi-set),
  `ToString`.
- **Inline CSS**: `Css` (get), `SetCss`, `SetCssMap` with deterministic
  `style` attribute parsing and serialization.
- **Data attributes**: `Data`, `DataOr`, `SetData`, `DataMap` with camelCase ⇄
  dashed key conversion.
- **Traversal**: `Contents`, `AddBack`, `Index`, `IndexSelector`,
  `IndexOfNode`, `AddSelection`, `AddNodes`, `NextUntil`, `PrevUntil`,
  `ParentsUntil`, `Even`, `Odd`. `End` now returns the previous stage of a
  traversal chain.
- **Selectors**: `:contains(text)`, `:has(sel)`, `:root`, `:parent`,
  `:only-of-type`, `:first-of-type`, `:last-of-type`, `:nth-of-type(n)`,
  `:nth-last-of-type(n)`, `:input`, `:checked`, `:selected`, `:enabled`,
  `:disabled`, `:header`; case-insensitive attribute matching (`[a=v i]`); and
  namespace-qualified type selectors (`svg|rect`, `*|title`).
- **Forms**: `SerializeArray` and `Serialize` (URL-encoded) with the standard
  successful-control rules.

### Changed

- The HTML parser now inserts an implicit `<tbody>` for table rows and cells
  written directly inside a `<table>`, matching browser tree construction.
- `Add` and `AddBack` now return their nodes in document order.

## [0.1.0]

- Initial release: HTML tokenizer, tree-construction parser, CSS selector
  engine, and a chainable jQuery-style Selection API.
