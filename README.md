# sval - Simple VALidator

**sval** - this is simple validator that can be fully configured by config file.

## Why sval?

sval supports the configuration of rules for specific structures and fields **entirely through json or yaml** configuration files. There is no need to use tags to set field validation rules because this can be done through a separate special file — `sval.yaml` or `sval.yml` or `sval.json`.

## Simple Example

There is our rules config file:

```yaml
// rules - main struct
rules:
  // user - our struct in code, info - embedded struct to user
  // name - field name in info struct
  user.info.name:
	// type - which rule set we want to use
    type: string
	// params - rules for this field
    params:
      required: true
      min_len: 10
      max_len: 25
      regex: "^[a-zA-Z ]+$"
  
  user.info.email:
	// another rule set - email
    type: email
    params:
      required: true
      rfc: true
      min_domain_len: 3
      excluded_domains:
        - temp.com
        - trashmail.org
        - example.com
      allowed_domains:
        - gmail.com
        - yahoo.com
```

There is our Go code for testing `sval`:

```go
type User struct {
	// we can name it as we want. But user.info - is good because
	// struct is User and embedded struct is Info
	Info Info `sval:"user.info"`
}

type Info struct {
	// There is no need to name tag like `user.info.name`,
	// because sval can detect rule set for this field already
	Name  string `sval:"name"`
	Email string `sval:"email"`
	Age   int    `sval:"age"`
	IP    string `sval:"ip"`
}

func main() {
	user := User{
		Info: Info{
			Name:  "John123123123",
			Email: "@test@examle.com",
		},
	}

	val, err := sval.New()
	if err != nil {
		panic(err)
	}

	if err := val.Validate(user); err != nil {
		fmt.Println("Validation error for user:", err)
	}
}
```

And we will see in console:

```
Validation error for user: {"errors":[{"id":1,"rule":"regex","rule_values":"^[a-zA-Z ]+$","message":"value does not match pattern"},{"id":2,"rule":"rfc","rule_values":true,"message":"email does not conform to RFC standards"},{"id":3,"rule":"allowed_domains","rule_values":["gmail.com","yahoo.com"],"message":"email domain is not allowed"}]}
```

Lets PRETTIFY this output a little bit:

```json
{
  "errors": [
    {
      "id": 1,
      "rule": "regex",
      "rule_values": "^[a-zA-Z ]+$",
      "message": "value does not match pattern"
    },
    {
      "id": 2,
      "rule": "rfc",
      "rule_values": true,
      "message": "email does not conform to RFC standards"
    },
    {
      "id": 3,
      "rule": "allowed_domains",
      "rule_values": [
        "gmail.com",
        "yahoo.com"
      ],
      "message": "email domain is not allowed"
    }
  ]
}
```

Yeah, this is json formatted as string.

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

	// number validations are not implemented yet
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
      "message": "value too short"
    },
    {
      "id": 2,
      "rule": "min_len",
      "rule_values": 5,
      "message": "value too short"
    },
    {
      "id": 3,
      "rule": "max_len",
      "rule_values": 10,
      "message": "value too long"
    }
  ]
}
```

P.S. in errors output `id` is not number of field or number of slice element. This is just a number XD

## Rule Sets

- [x] String
- [x] Email
- [ ] IP
- [ ] Number
- [ ] Mac
- [ ] AND MORE

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