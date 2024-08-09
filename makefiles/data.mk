DATA += data/blackmatrix7/Advertising.list
DATA += data/blackmatrix7/ChinaMax.list
DATA += data/blackmatrix7/Claude.list
DATA += data/blackmatrix7/Copilot.list
DATA += data/blackmatrix7/Developer.list
DATA += data/blackmatrix7/Download.list
DATA += data/blackmatrix7/Gemini.list
DATA += data/blackmatrix7/GlobalMedia.list
DATA += data/blackmatrix7/Lan.list
DATA += data/blackmatrix7/NTPService.list
DATA += data/blackmatrix7/OneDrive.list
DATA += data/blackmatrix7/OpenAI.list
DATA += data/DustinWin/geoip-all.db
DATA += data/DustinWin/geosite-all.db
DATA += data/MetaCubeX/geoip.db
DATA += data/MetaCubeX/geosite-lite.db
DATA += data/MetaCubeX/geosite.db
DATA += data/NotSFC/Emby.json

.PHONY: data
data: $(DATA)

data/blackmatrix7/%.list:
	@ mkdir --parents --verbose "$(@D)"
	wget --output-document="$@" "https://github.com/blackmatrix7/ios_rule_script/raw/master/rule/Clash/$*/$*.list"

data/DustinWin/%.db:
	@ mkdir --parents --verbose "$(@D)"
	wget --output-document="$@" "https://github.com/DustinWin/ruleset_geodata/releases/download/sing-box/$*.db"

data/MetaCubeX/%.db:
	@ mkdir --parents --verbose "$(@D)"
	wget --output-document="$@" "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/$*.db"

data/NotSFC/%.json:
	@ mkdir --parents --verbose "$(@D)"
	wget --output-document="$@" "https://github.com/NotSFC/rulelist/raw/main/sing-box/$*/$*.json"
