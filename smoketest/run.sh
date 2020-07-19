#!/bin/sh

alias curl="curl --fail --silent"
alias jq="jq --exit-status"

extract='map({
	"logicalId": .logicalId,
	"exact": (.relevance == 1.0),
	"matchingAlias": .matchingAlias,
	"otherAliases": (.otherAliases | sort_by(.))
})'

search() {
	name=$1
	assertion=$2

	response=$(curl --get ${API}/search  --data-urlencode "name=$name")
	if [[ $? == 0 ]] && $(echo ${response} | jq "${extract} | ${assertion}" > /dev/null); then
		echo "PASS | search: ${name}"
	else
		echo "FAIL | search: ${name}"
		cat <<- EOF | sed 's/^/     | /g'
			expect: ${assertion}
			actual: $(echo $response | jq "$extract")
		EOF
	fi
	echo
}

until $(curl --output /dev/null ${API}/status); do
	echo 'waiting for api...'
	sleep 5
done

search 'KCST' '.[0] ==
	{
		"logicalId": 7212,
		"exact": true,
		"matchingAlias": "KCST",
		"otherAliases": [
			"Committee for Space Technology",
			"DPRK Committee for Space Technology",
			"Department of Space Technology of the DPRK",
			"Korean Committee for Space Technology"
		]
	}'

search 'Iran Modern Devices' '.[0] ==
	{
		"logicalId": 6446,
		"exact": false,
		"matchingAlias": "Aran Modern Devices",
		"otherAliases": ["AMD"]
	}'

search 'abou ali' '.[0:2] |
	contains([
		{
			"logicalId": 13,
			"exact": true,
			"matchingAlias": "Abou Ali",
			"otherAliases": [
				"Abu Ali",
				"Saddam Hussein Al-Tikriti"
			]
		}, {
			"logicalId": 1095,
			"exact": true,
			"matchingAlias": "Abou Ali",
			"otherAliases": [
				"Faycal",
				"Noureddine Ben Ali Ben Belkassem Al-Drissi",
				"Noureddine Drissi"
			]
		}
	])'

search 'Pop Credit Bank' '.[0] ==
	{
		"logicalId": 6897,
		"exact": false,
		"matchingAlias": "Popular Credit Bank",
		"otherAliases": ["Banca Populară de Credit"]
	}'

search 'Université Malek Ashtar' '.[0] ==
	{
		"logicalId": 5286,
		"exact": true,
		"matchingAlias": "Université Malek Ashtar",
		"otherAliases": [
			"Malek Ashtar University",
			"Malek Ashtar universitetas",
			"Malek Ashtar-Universität",
			"Universidade Malek Ashtar",
			"Universitatea Malek Ashtar",
			"Università Malek Ashtar",
			"Università Malek Ashtar",
			"Univerza za obrambno tehnologijo Malek Ashtar",
			"Univerzita Malek Ashtar",
			"Univerzita Maleka Aštara",
			"Πανεπιστήμιο Malek Ashtar",
			"Университет Malek Ashtar"
		]
	}'

search 'οργάνωση Al-Qaida στην Αραβική Χερσόνησο' '.[0] |
	.logicalId == 5523 and
	.exact == true and
	.matchingAlias == "οργάνωση Al-Qaida στην Αραβική Χερσόνησο" and
	(.otherAliases | contains([
		"AAS",
		"Al-Kaida na Półwyspie Arabskim",
		"Al-Kaida organizacije Džihad z Arabskega polotoka,",
		"Al-Kaida w południowej części Półwyspu Arabskiego",
		"Al-Kaida z južnega dela Arabskega polotoka",
		"Al-Qa'"'"'ida van het zuidelijk Arabisch Schiereiland",
		"Al-Qa'"'"'ida-organisatie van het Arabisch Schiereiland",
		"Al-Qa‘ida van het Arabisch Schiereiland",
		"Al-Qaeda en la península arábiga",
		"Al-Qaid a de l’organisation du Djihad dans la péninsule arabique",
		"Al-Qaida dans la péninsule arabique",
		"Al-Qaida din Peninsula Arabia de Sud",
		"Al-Qaida din Yemen",
		"Al-Qaida fil-Peniżola Għarbija",
		"Al-Qaida i den sydlige del af Den Arabiske Halvø",
		"Al-Qaida in the Arabian Peninsula",
		"Al-Qaida in Yemen",
		"Al-Qaida Organization in the Arabian Peninsula",
		"Al-Qaida på Den Arabiske Halvø",
		"Al-Qaida tal-Jihad Organization fil-Peniżola Għarbija,",
		"Al-Qaida στην Νότια Αραβική Χερσόνησο",
		"Al-Qaida της οργάνωσης Jihad στην Αραβική Χερσόνησο",
		"AQAP",
		"AQPA",
		"AQY",
		"Organizacja dżihadu na Półwyspie arabskim Al-Kaida,",
		"Organizația Al-Qaida din Peninsula Arabia",
		"Tanzim Qa'"'"'idat al-Jihad fi Jazirat al-Arabm"
	]))'


exit 0
