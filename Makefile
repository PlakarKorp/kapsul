GO =		go

DESTDIR =
PREFIX =	/usr/local
BINDIR =	${PREFIX}/bin
MANDIR =	${PREFIX}/man

INSTALL =	install
INSTALL_PROGRAM=${INSTALL} -m 0555
INSTALL_MAN =	${INSTALL} -m 0444

all: kapsul

kapsul:
	${GO} build -v .

install:
	mkdir -p ${DESTDIR}${BINDIR}
	mkdir -p ${DESTDIR}${MANDIR}/man1
	${INSTALL_PROGRAM} kapsul ${DESTDIR}${BINDIR}
	${INSTALL_MAN} kapsul.1 ${DESTDIR}${MANDIR}/man1

clean:
	rm kapsul

.PHONY: all kloset install clean
