package cascade

func main() {

	c := Container()
	c.Register("rpc", &RPC)
	c.Register("jobs", &Jobs)
}
