# Scenery
Scenery is a zero dependencies CLI tool to prettify `terraform plan` outputs to be easier to read and digest. A lot of inspiration was drawn from [Terraform Landscape](https://github.com/coinbase/terraform-landscape).

<p align="center">
  <img src="https://s3.amazonaws.com/scenery-public-assets/scenery_recording.svg">
</p>

## Installing

If you have a functional Go environment, you can install `scenery` with the following command:

```bash
$ go get -u github.com/dmlittle/scenery
```

## Usage

```bash
$ terraform plan ... | scenery
```

If you wish to suppress the color output you may pass a `--no-color` flag to `scenery`.
```bash
$ terraform plan ... | scenery --no-color
```

## License

The MIT License (MIT) - see [`LICENSE.md`](https://github.com/dmlittle/scenery/blob/master/LICENSE.md) for more details.
