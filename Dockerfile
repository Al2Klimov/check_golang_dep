FROM golang:1.11 as build

RUN go get github.com/golang/dep \
	&& go install github.com/golang/dep/...

ADD . /go/src/github.com/Al2Klimov/check_golang_dep

RUN cd /go/src/github.com/Al2Klimov/check_golang_dep \
	&& /go/bin/dep ensure \
	&& go generate \
	&& go install .

FROM grandmaster/check-plugins-demo

SHELL ["/bin/bash", "-exo", "pipefail", "-c"]

RUN apt-get update ;\
	DEBIAN_FRONTEND=noninteractive apt-get install --no-install-{recommends,suggests} -y \
		golang-go git-core ;\
	apt-get clean ;\
	rm -vrf /var/lib/apt/lists/*

RUN a2enmod cgi ;\
	a2enmod alias ;\
	a2enmod env ;\
	for cat in lolcat grumpycat; do \
		mkdir -p /var/www/git/toni/$cat.git ;\
		pushd /var/www/git/toni/$cat.git ;\
		git init --bare ;\
		git config http.receivepack true ;\
		git config http.uploadpack true ;\
		popd ;\
	done ;\
	chown -R www-data:www-data /var/www/git

COPY _docker/apache2-git.conf /etc/apache2/sites-available/git.conf
RUN a2ensite git

COPY _docker/git-toni.sh /
COPY _docker/supervisord.conf /etc/supervisord/conf.d/git-toni.conf

COPY --from=build /go/bin/dep /usr/local/bin/
COPY --from=build /go/bin/check_golang_dep /usr/lib/nagios/plugins/
COPY icinga2/check_golang_dep.conf _docker/icinga2-monobjs.conf /etc/icinga2/conf.d/
