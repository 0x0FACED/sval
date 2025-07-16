# sval

<div align="center">

🎯 Simple yet powerful validation library for Go with configuration files support

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Status](https://img.shields.io/badge/status-beta-yellow)]()

</div>

## 🌟 Features

- Configuration-driven validation using YAML/JSON files
- Support for nested structures and slices
- Multiple email validation strategies (RFC 5321, RFC 5322, HTML5)
- Extensible rule set system
- Clear and detailed validation errors
- Zero external dependencies (except testing)

## 📦 Installation

```bash
go get github.com/0x0FACED/sval
```

## 🚀 Quick Start

### 1. Create configuration file (`sval.yaml`)

```yaml
rules:
  user.info.name:
    type: string
    params:
      required: true
      min_len: 10
      max_len: 25
      regex: "^[a-zA-Z ]+$"
  
  user.info.email:
    type: email
    params:
      required: true
      strategy: rfc5322
      allowed_domains:
        - gmail.com
        - yahoo.com
```

### 2. Define your structures

```go
type User struct {
    Info Info `sval:"user.info"`
}

type Info struct {
    Name  string `sval:"name"`
    Email string `sval:"email"`
}
```

### 3. Validate your data

```go
func main() {
    user := User{
        Info: Info{
            Name:  "John Doe",
            Email: "@john@example.com",
        },
    }

    validator, err := sval.New()
    if err != nil {
        panic(err)
    }

    if err := validator.Validate(user); err != nil {
        fmt.Println(err)
    }
}
```

And we will see in console:

```
{"errors":[{"id":1,"rule":"min_len","rule_values":10,"provided":"John Doe","message":"value too short"},{"id":2,"rule":"strategy","rule_values":"rfc5322","provided":"@john@example.com","message":"email does not conform to chosen strategy"},{"id":3,"rule":"allowed_domains","rule_values":["gmail.com","yahoo.com"],"provided":"@john@example.com","message":"email domain is not allowed"}]}
```

Lets PRETTIFY this output a little bit:

```json
{
  "errors": [
    {
      "id": 1,
      "rule": "min_len",
      "rule_values": 10,
      "provided": "John Doe",
      "message": "value too short"
    },
    {
      "id": 2,
      "rule": "strategy",
      "rule_values": "rfc5322",
      "provided": "@john@example.com",
      "message": "email does not conform to chosen strategy"
    },
    {
      "id": 3,
      "rule": "allowed_domains",
      "rule_values": [
        "gmail.com",
        "yahoo.com"
      ],
      "provided": "@john@example.com",
      "message": "email domain is not allowed"
    }
  ]
}
```

## 📖 Configuration Reference

### Rule Structure

Each validation rule in `sval.yaml` consists of:

- **Path**: Field path using dot notation (e.g., `user.info.email`)
- **Type**: Validation type (`email`, `string`, `number`, `ip`, `password`)
- **Params**: Type-specific configuration parameters

### Available Validation Types

#### 1. Email
```yaml
type: email
params:
  required: true
  strategy: rfc5322  # Available: rfc5321, rfc5322, html
  allowed_domains:   # Optional
    - example.com
    - company.com
```

#### 2. String
```yaml
type: string
params:
  required: true
  min_length: 2
  max_length: 50
  pattern: "^[a-zA-Z0-9]+$"  # Optional regex pattern
```

#### 3. Int
```yaml
type: int
params:
  required: true
  min: 0
  max: 100
```

#### 4. Float
```yaml
type: float
params:
  required: true
  min: 0.0
  max: 100.0
```

#### 5. IP Address

(not impl)

```yaml
type: ip
params:
  required: true
  version: both  # ipv4, ipv6, or both
```

#### 6. Password

(not impl)

```yaml
type: password
params:
  required: true
  min_length: 8
  require_uppercase: true
  require_lowercase: true
  require_numbers: true
  require_special: true
  special_chars: "!@#$%^&*"  # Optional custom set
```

## 🔍 Error Handling

Validation errors are returned as structured JSON:

```go
type ValidationError struct {
    Errors []struct {
        ID         int    `json:"id"`
        Rule       string `json:"rule"`
        RuleValues any    `json:"rule_values"`
        Provided   any    `json:"provided"`
        Message    string `json:"message"`
    } `json:"errors"`
}
```

### Error Example

```json
{
  "errors": [
    {
      "id": 1,
      "rule": "min_len",
      "rule_values": 10,
      "provided": "John Doe",
      "message": "value too short"
    },
    {
      "id": 2,
      "rule": "strategy",
      "rule_values": "rfc5322",
      "provided": "@john@example.com",
      "message": "email does not conform to chosen strategy"
    },
    {
      "id": 3,
      "rule": "allowed_domains",
      "rule_values": [
        "gmail.com",
        "yahoo.com"
      ],
      "provided": "@john@example.com",
      "message": "email domain is not allowed"
    }
  ]
}
```

## 🔧 Custom Rules

(not impl) - will be in the future.

You can extend `sval` with custom validation rules:

1. Create your rule type:
```go
type CustomRule struct{}

func (r *CustomRule) Validate(value interface{}, params map[string]interface{}) error {
    // Your validation logic here
    return nil
}
```

2. Register the rule:
```go
validator.RegisterRule("custom", &CustomRule{})
```

3. Use in configuration:
```yaml
rules:
  user.field:
    type: custom
    params:
      your_param: value
```

## Another example with slices

sval supports SLICES and diving into elements to validate them all!

```yaml
rules:
  // use [] to tell sval that we want to dive into slice
  orders[].order_id:
    type: string
    params:
      required: true
      min_len: 5
      max_len: 10
      regex: "^[A-Z0-9]+$"
```

```go
type OrderInfo struct {
	OrderID string `sval:"order_id"`
	Amount  int    `sval:"amount"`
}

type Order struct {
	Orders []OrderInfo `sval:"orders"`
}

func main() {
	val, err := sval.New()
	if err != nil {
		panic(err)
	}

	order := Order{
		Orders: []OrderInfo{
			{OrderID: "123", Amount: 100},
			{OrderID: "SHO", Amount: 200},
			{OrderID: "TOOLONGID1234567890", Amount: 300},
			{OrderID: "VALID123", Amount: 10000},
		},
	}

	if err := val.Validate(order); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Order validation passed")
	}
}
```

Output:

```json
{
  "errors": [
    {
      "id": 1,
      "rule": "min_len",
      "rule_values": 5,
      "provided": "123",
      "message": "value too short"
    },
    {
      "id": 2,
      "rule": "min_len",
      "rule_values": 5,
      "provided": "SHO",
      "message": "value too short"
    },
    {
      "id": 3,
      "rule": "max_len",
      "rule_values": 10,
      "provided": "TOOLONGID1234567890",
      "message": "value too long"
    }
  ]
}
```

P.S. in errors output `id` is not number of field or number of slice element. This is just a number XD

## Documentation

### File Format

`sval` supports struct validating with embedded structs and slices of structs.

Config file must be named `sval.yaml` or `sval.yml` or `sval.json` and placed in the root of working directory. Prefer name `sval.yaml`.

### Simple config

```yaml
rules:
  user.name:
    type: string
    params:
      required: true
      min_len: 5
      max_len: 16
      regex: "custom regex"
      alphanum: true
```

1. `rules` must be first element
2. After `rules` block, validation rules are defined.
3. To define validation rules for field in struct, you must enter `structName.fieldName`, `user.name` in the example below.
4. You must to define type of validation rule: `string` in the example.
5. Next you define your rules in the `params` section.
6. In the code you must add tag for fields in the struct:
```go
type User struct {
    Name string `sval:user.name`
}
```

Tag `sval` shows validator that this field must be validated. `user.name` tells sval that he must use rule `user.name` from config file.

### Allowed rule types

`sval` currently supported next types:
1. `string`
2. `int`
3. `float`
4. `email`

### Base rules

Base Rules include `required` only supported by all rule sets.

### string type

String type can define next rules:
1. `required` - from Base Rules
2. `min_len` - minimal length of string in chars, not bytes
3. `max_len` - max length of string in chars, not bytes
4. `regex` - custom regex
5. `alphanum` - not implemented yet

### email rules

Email type can define next rules:
1. `required`
2. `strategy` - accepts 3 values: `rfc5322`, `rfc5321`, `html` First - use rfc5322 validation, second - validation for SMTP, third - just uses regex like html form `type=email`
3. `min_domain_len` - not recommended to use
4. `excluded_domains` - list of excluded domains. If you set domain `example.com` as excluded, than emails like `johndoe@example.com`, `alisathefuture@example.com` etc will be validated as incorrect.
5. `allowed_domains` - list of allowed domains. The opposite effect to `excluded_domains`.
6. `regex`

## Rule Sets

- [x] String
- [x] Email
- [ ] IP
- [x] Int
- [x] Float
- [ ] Mac
- [ ] Date and time
- [ ] geo
- [ ] url
- [ ] colors
- [ ] iban
- [ ] isbn
- [ ] arrays
- [ ] phones
- [ ] uuids
- [ ] custom validations

## A little roadmap

- [ ] Add cli that can validate config file
- [ ] Add codegen go code from config file
- [ ] Add more rule sets
- [ ] Add more rules
- [ ] Add NORMAL id in `ValidationError`
- [ ] Add more fields to `ValidationError`
- [ ] Add custom rules support
- [ ] Add tests for 100% coverage
- [ ] Add good readable documentation