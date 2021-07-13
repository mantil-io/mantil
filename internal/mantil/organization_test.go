package mantil

// func setDevEnvPaths(t *testing.T) {
// 	var err error
// 	templatesFolder, err = filepath.Abs("../../templates")
// 	require.NoError(t, err)
// 	modulesFolder, err = filepath.Abs("../../..")
// 	require.NoError(t, err)
// 	secretsFolder, err = filepath.Abs("../../../../secrets")
// 	require.NoError(t, err)

// 	err = shell.PrepareHome(rootFolder+"/home", secretsFolder)
// 	require.NoError(t, err)
// }

// func TestOrganizationLoad(t *testing.T) {
// 	setDevEnvPaths(t)

// 	org := Organization{}
// 	err := org.Load("org1")
// 	require.NoError(t, err)

// 	require.Equal(t, org.Name, "org1")
// 	require.Equal(t, org.DNSZone, "org1.mantil.team")
// 	require.Equal(t, org.FunctionsBucket, "org1.mantil.team-lambda-functions")

// 	var cert Cert
// 	err = org.LoadProject("cert", &cert)
// 	require.NoError(t, err)
// 	require.Equal(t, cert.Arn("app1"), "arn:aws:acm:us-east-1:784096511694:certificate/0adc7cd5-ee6a-4bab-ba51-f50d68269d4e")
// }

// func TestCertPrepare(t *testing.T) {
// 	setDevEnvPaths(t)

// 	var c Cert
// 	c.testData()
// 	c.Name = "cert1"
// 	err := c.Organization.PrepareProject("cert", c.Name, c)
// 	require.NoError(t, err)
// }

// func TestSpaApply(t *testing.T) {
// 	setDevEnvPaths(t)

// 	org := Organization{}
// 	err := org.Load("org1")
// 	require.NoError(t, err)

// 	var spa Spa
// 	err = org.LoadProject("app1", &spa)
// 	require.NoError(t, err)

// 	spa.Organization = org
// 	err = spa.Apply()
// 	require.NoError(t, err)
// }

// func pp(o interface{}) {
// 	buf, _ := json.MarshalIndent(o, "  ", "  ")
// 	fmt.Printf("%s\n", buf)
// }

// func TestSpaTemplate(t *testing.T) {
// 	setDevEnvPaths(t)

// 	var spa Spa
// 	spa.testData()
// 	pp(spa)

// 	content, err := template.Render(templatesFolder+"/spa/main.tf", spa)
// 	require.NoError(t, err)
// 	fmt.Printf("%s\n", content)
// }
