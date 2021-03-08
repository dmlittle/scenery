# ⚠️  **This project is no longer maintained** ⚠️

Scenery was specifically designed to parse the plan output of Terraform 0.11 which has since been [deprecated](https://www.hashicorp.com/blog/deprecating-terraform-0-11-support-in-terraform-providers). Anyone considering building upon this tool is recommended to look into the new terraform plan [JSON output](https://www.terraform.io/docs/internals/json-format.html) (introdcued in Terrafrom 0.12) rather than parsing raw text output. Additionally Terraform 0.14 introduced [concise diff plan outputs](https://www.hashicorp.com/blog/terraform-0-14-adds-a-new-concise-diff-format-to-terraform-plans) that does most of what scenery does today. 

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
