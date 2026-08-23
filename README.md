# nodeset

Parse, expand, split, and fold node expressions such as `a[01-02]r[1-4]n[01-20]`.
The package also accepts bounded IPv4/IPv6 bracket ranges and CIDR expressions.

The parser keeps leading-zero width, returns deterministic natural ordering, and
rejects malformed members instead of silently accepting an empty node. `Add` is
transactional: a failed update does not partially mutate an existing `NodeSet`.

The package is syntax-focused and does not resolve DNS or validate host
reachability. Callers that use the values as shell targets must apply their own
host allowlist before execution.

Public helpers:

- `NewNodeSet`, `Expand`, and `Yield` for parsing and iteration;
- `Merge` for canonical aggregation and `ExpandCIDR` for one CIDR network;
- `Split` for balanced worker partitions; non-positive widths return no parts.

The project was originally derived from the Grendel nodeset implementation:
https://github.com/ubccr/grendel/nodeset
