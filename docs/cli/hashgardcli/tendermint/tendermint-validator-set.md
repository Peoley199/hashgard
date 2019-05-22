# hashgardcli tendermint tendermint-validator-set

## Description

Get the full tendermint validator set at given height

## Usage

```
  hashgardcli tendermint tendermint-validator-set [height] [flags]
```

## Flags

| Name, shorthand  | Default               | Description            | Required               |
| ------------ | --------------------- | -------------------------- | ---------------------- |
| --chain-id   | No                    | Chain ID of Tendermint node                     | No|
| --node       | tcp://localhost:26657 | Node to connect to                      | No        |
| --trust-node | true        | Trust connected full node (don't verify proofs for responses)| No |

**Global flags, query command flags** [hashgardcli](../README.md)

## Example

```shell
hashgardcli tendermint tendermint-validator-set 114360 --trust-node
```

The result is as follows：

```
block height: 123

  Address:          gardvalcons13ja77lpt0deamvuwz5eugy9kwkutxukjwjwwf3
  Pubkey:           gardvalconspub1zcjduepqgsmuj0qallsw79hjj9qztcke6hj3ujdcpjv249uny9fvzp4eulms0tqvgs
  ProposerPriority: 0
  VotingPower:      1000

```
