SING_BOX ?= sing-box

OUTPUTS += output/geoip/cn.srs
OUTPUTS += output/geosite/ads.srs
OUTPUTS += output/geosite/cn.srs
OUTPUTS += output/geosite/private.srs
OUTPUTS += output/geosite/proxy.srs
OUTPUTS += output/rule-set/ads.srs
OUTPUTS += output/rule-set/ai.srs
OUTPUTS += output/rule-set/cn.srs
OUTPUTS += output/rule-set/download.srs
OUTPUTS += output/rule-set/emby.srs
OUTPUTS += output/rule-set/media.srs
OUTPUTS += output/rule-set/private.srs
OUTPUTS += output/rule-set/proxy.srs

.PHONY: output
output: output/README.md $(OUTPUTS) $(OUTPUTS:.srs=.json)

output/README.md $(OUTPUTS:.srs=.json) &: scripts/build.py $(DATA)
	python "$<"
	prettier --write --ignore-path "" "output/README.md"

output/%.srs: output/%.json
	$(SING_BOX) rule-set compile "$<" --output "$@"
