.PHONY: update-secrets

update-secrets:
	tar cvf secrets.tar .prod.env .stage.env client-secret.json sqreen.yaml
	travis encrypt-file secrets.tar --add --com