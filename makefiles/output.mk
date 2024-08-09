SING_BOX ?= sing-box

OUTPUTS += output/geosite/ads.srs
OUTPUTS += output/geosite/cn.srs
OUTPUTS += output/geosite/private.srs
OUTPUTS += output/rule-set/ads.srs
OUTPUTS += output/rule-set/cn.srs
OUTPUTS += output/rule-set/download.srs
OUTPUTS += output/rule-set/emby.srs
OUTPUTS += output/rule-set/media.srs
OUTPUTS += output/rule-set/private.srs

.PHONY: output
output: output/README.md $(OUTPUTS) $(OUTPUTS:.srs=.json)

output/README.md: scripts/summary.py $(OUTPUTS:.srs=.json)
	python "$<" > "$@"
	prettier --write --ignore-path "" "$@"

output/rule-set/%.srs: output/rule-set/%.json
	$(SING_BOX) rule-set compile "$<" --output "$@"

output/geosite/%.srs: output/geosite/%.json
	$(SING_BOX) rule-set compile "$<" --output "$@"

output/rule-set/%.json: preset/%.py $(DATA)
	python "$<"

output/geosite/%.json: preset/geosite.py output/rule-set/%.json
	python "$<" "$(word 2,$^)" "$@"

output/rule-set/ai.json: output/rule-set/ads.json
output/rule-set/cn.json: output/rule-set/ads.json output/rule-set/private.json
output/rule-set/download.json: output/rule-set/ads.json output/rule-set/cn.json
output/rule-set/media.json: output/rule-set/ads.json
output/rule-set/private.json: output/rule-set/ads.json
