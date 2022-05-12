package register

var suite = map[string]NewSuite{}

func AddSuite(name string, test NewSuite) {
	suite[name] = test
}

func GetSuites() map[string]NewSuite {
	return suite
}
