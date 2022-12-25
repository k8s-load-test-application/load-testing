CASES := $(shell find ./cmd -name '*.js' -exec basename '{}' ';' |       \
                        tr '\n' '\0' |                                   \
                        xargs -0 -n 1 |                                  \
                        awk '{ split($$0,a,".js"); print a[1]".out" }')


DURATION := 60
EMAIL    := $(shell git log --format='%ae' HEAD^!)
VUS      := 200

%.out: %.js
	@-k6 run --vus ${VUS}                                           \
	         --duration ${DURATION}s - <$< | tee -a $@ ||        	\
	         mail -s "🚨 Failed performance test" ${EMAIL} <<< "$(shell date '+%D %T') - Test $< fails"

all: $(CASES)

clean:
	@-find . -name '*.out' -delete
