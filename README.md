# Protor
library for forwarding data to prometheus aggregator for internal monitoring

#Interface input and output

````
type ProtorInterface interface {
	Decode(context.Context, io.Reader) ([]*pt.ProtorData, error) //decode intejson input to match go struct
	Encode(context.Context, *pt.ProtorData) string               //encode go json data to match protor standard
	Valid(context.Context, *pt.ProtorData) bool                  //Validate data
	Work(context.Context, *pt.ProtorData) error                  //send data to protor
}
````
