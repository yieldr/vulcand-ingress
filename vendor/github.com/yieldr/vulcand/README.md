# Vulcand

Yieldr's fork of Mailgun's [vulcand](https://github.com/vulcand/vulcand) built using [`vbundle`](http://vulcand.github.io/middlewares.html#vbundle) to add custom middleware.

# Usage

## Kubernetes

Typically this project is deployed alongside [vulcand-ingress](https://github.com/yieldr/vulcand-ingress) or [romulus](https://github.com/albertrdixon/romulus) as a [Kubernetes](https://kubernetes.io) [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-controllers). See the referring projects for their specific usage examples.

## Development

Build the binaries using `make`.

	make build OS=linux

Then start docker compose

	docker-compose up -d

When all services are up and running, you'll need to configure vulcan to route traffic to the upstream servers. A dummy upstream is included for convenience that simply displays the nginx default web page.

### Frontends, backends  and servers

The following configuration will create a new upstream backend (`nginx`), a server (`nginx-srv`) and a new frontend (`nginx`).

	vctl backend upsert -id nginx
	vctl server upsert -b nginx -id nginx-srv -url http://nginx:80
	vctl frontend upsert -id nginx -b nginx -route 'PathRegexp("/.*")'

### Middleware

This command will add `oauth2` middleware to the `nginx` frontend.

	vctl oauth2 upsert --id nginx-oauth --frontend nginx \
		--issuerUrl $ISSUER_URL \
		--clientId $CLIENT_ID \
		--clientSecret $CLIENT_SECRET \
		--redirectUri $REDIRECT_URI


vctl backend upsert -id nginx
vctl server upsert -b nginx -id nginx-srv -url http://nginx:80
vctl frontend upsert -id nginx -b nginx -route 'PathRegexp("/.*")'
vctl oauth2 upsert --id nginx-oauth --frontend nginx \
	--issuerUrl https://yieldr.eu.auth0.com \
	--clientId JklNORC4LOPSotjX25sZVcam6ZWpM53f \
	--clientSecret Pu0kl30h4ut7pFh5baczOhLlyCpBv-pm9iQOKFsVEsVdeUgEGlh4RY0zeknl4oUx \
	--redirectUri http://localhost:8181/callback
