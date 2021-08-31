package git

const (
	DeployWorkflow = `
name: Mantil deploy
on: [push, workflow_dispatch]
jobs:
  Deploy:
    runs-on: ubuntu-latest
    env:
      MANTIL_TOKEN: ${{ secrets.MANTIL_TOKEN }}
      MANTIL_BACKEND_URL: ${{ secrets.MANTIL_BACKEND_URL }}
    steps:
      - name: Checkout repository
        uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '^1.16'
      - name: Deploy
        run: |
          wget -q https://mantil-downloads.s3.eu-central-1.amazonaws.com/mantil
          chmod +x mantil
          ./mantil deploy
`
)
