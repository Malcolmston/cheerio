# API deviations from cheerio

This port aims for behavioural parity with `cheerio@1.0.0` (verified by the
cross-language parity harness in the aggregator repo, `parity/cheerio/`). The
cases below are the handful of places where the Go port deliberately differs.
Each corresponds to a parity case marked `"deviation"`, which the harness counts
separately from a bug.

## Accessors return `""`, not `undefined`

- **`Selection.Css(prop)`** returns `""` for a property that is not set.
- **`Selection.Val()`** returns `""` for a control that has no submittable value
  (e.g. `<input type="file">`).

cheerio returns JavaScript `undefined` in both cases. The Go port uses
string-only accessors by design (its own test suite locks this in), so an absent
value is the empty string. Callers that must distinguish "absent" from "empty"
can inspect the underlying attribute directly.

_Parity cases: `acc-css-absent`, `acc-val-no-value-attr`._

## `multiple` select serialisation submits the option value, not its text

For `<option value="m1" selected>M1</option>` inside a `<select multiple>`,
cheerio's `serialize()` / `serializeArray()` submit the option's **text**
(`M1`). Per the HTML form-submission algorithm the submitted value of an option
with a `value` attribute is that attribute (`m1`), which is also what cheerio's
own single-select path returns. The Go port submits `value` consistently, so it
disagrees with upstream only for `multiple` selects — upstream is the outlier
here (a bug in `Cheerio.prototype.val` for `select[multiple]`).

_Parity cases: `form-serialize`, `form-serialize-array`,
`form-serialize-select-multiple`._

## `Selection.RemoveData` exists

`removeData` is not part of `Cheerio.prototype` in `cheerio@1.0.0`, so calling it
upstream throws. The Go port keeps `RemoveData` as an extra convenience method
(it clears both the in-memory data cache and any `data-*` attributes).

_Parity case: `man-removedata`._

## Namespace-qualified type selectors are supported

`css-select` throws _"Namespaced tag names are not yet supported"_ for a selector
like `*|p`. The Go port accepts it (matching on the local name), so it is a
superset of upstream here.

_Parity case: `sel-namespace-qualified`._

## The doctype is preserved

parse5's fragment-parsing mode (used by the harness's Node runner) drops a
leading `<!DOCTYPE html>`. The Go port has no fragment/document switch and always
preserves the doctype, matching upstream's *document* mode. This is a
parsing-mode artefact of the comparison rather than a behavioural bug.

_Parity cases: `parse-comments-doctype`, `ser-doctype-preserved`._

## Known limitation (not a deviation)

The parser does not implement the HTML5 **adoption agency algorithm** /
active-formatting-element reconstruction, so deeply mis-nested formatting tags
such as `<b>1<i>2</b>3</i>` are not re-opened the way a full HTML5 parser would.
These remain genuine differences from upstream (parity cases
`parse-misnested-format`, `parse-misnested-deep`) and are tracked as a real gap,
not marked as deviations.
