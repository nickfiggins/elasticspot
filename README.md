
# ElasticSpot
Plug-in lambda function for reassigning Elastic IPs to new spot instances on termination inside an ECS cluster.



## Usage/Examples

```go
package main

import (
	"github.com/nickfiggins/elasticspot"
	...
)


...

func main() {
	lambda.Start(elasticspot.NewHandler(ec2_sess, elasticIp).Handle)
}
```

## Contributing

Contributions are always welcome!


## License

[MIT](https://choosealicense.com/licenses/mit/)