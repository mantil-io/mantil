package github

const (
	DeployWorkflow = `
name: Mantil deploy
on: [push, workflow_dispatch]
jobs:
  Deploy:
    runs-on: ubuntu-latest
    env:
      MANTIL_TOKEN: ${{ secrets.MANTIL_TOKEN }}
    steps:
      - name: Checkout repository
        uses: actions/checkout@v2
      - name: Setup terraform
        uses: hashicorp/setup-terraform@v1
      - name: Configure aws credentials
        uses: aws-actions/configure-aws-credentials@v1
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: eu-central-1
      - name: Deploy
        run: |
          wget -q https://mantil-downloads.s3.eu-central-1.amazonaws.com/mantil
          chmod +x mantil
          ./mantil deploy
`
)
