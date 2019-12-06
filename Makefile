.PHONY: update-secrets

update-secrets:
	tar cvf secrets.tar .prod.env .stage.env gcp-deploy.json firebase.stage.json firebase.prod.json sqreen.yaml
	travis encrypt-file secrets.tar --add --com