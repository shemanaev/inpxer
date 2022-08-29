default:

test:
	go test -short ./...

xspreak:
	xspreak -D ./ -o internal/i18n/locale/inpxer.pot -t "ui/templates/*.gohtml"
