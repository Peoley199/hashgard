# hashgardcli distribution commission


## Description

Query validator commission rewards from delegators to that validator:
## Usage

```
hashgardcli distribution commission [validator] [flags]
```

## Flags

**Global flags, query command flags** [hashgardcli](../README.md)


## Example

```
hashgardcli distribution commission gardvaloper1m0g2n0r7l6s44sac2knmx40hlsdyv4esgcwg8w -o=json --trust-node
```

The result is as follows：

```
[
 {
  "denom": "gard",
  "amount": "0.337966901187138531"
 }
]
```
