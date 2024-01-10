package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fuzzSave struct {
	ValSet  *ValidatorSet `json:"vs"`
	ChainID string        `json:"cid"`
	BlockID BlockID       `json:"bid"`
	Height  int64         `json:"h"`
	Commit  *Commit       `json:"comm"`
}

func FuzzValSetVerifyCommit(f *testing.F) {
	if testing.Short() {
		f.Skip("Running in -short mode")
	}

	seeds := []string{
		`{"vs":{"validators":[{"address":"2224569C1EC8B6B49534E01E82BB1CF514D33F5E","pub_key":"L2lD6cR6jNq+UM2XaS/bwP3NqImFw8zYd01FnODmQNg=","voting_power":10,"proposer_priority":0}],"proposer":{"address":"2224569C1EC8B6B49534E01E82BB1CF514D33F5E","pub_key":"L2lD6cR6jNq+UM2XaS/bwP3NqImFw8zYd01FnODmQNg=","voting_power":10,"proposer_priority":0}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":2,"validator_address":"2224569C1EC8B6B49534E01E82BB1CF514D33F5E","timestamp":"2023-12-13T23:46:39.357412-08:00","signature":"ufBxQXJqeIisMOYjJc6ekIcvTlGxvSVi39PZJ2kZQ0HShk5rBLKJL4wXfeWT/is2UgbYqkvLvVdrLNE4S+78Dw=="}]}}`,

		`{"vs":{"validators":[{"address":"33D285936DEDAEB03B9497594130FEEADAFD566F","pub_key":"YdZcuYQh1nFOocmkbP+itvDFS9ZbpZkW4RFuyga9BX8=","voting_power":10,"proposer_priority":-80},{"address":"43C38459CE6D21F3CED9DA4F483E1E5779ACE736","pub_key":"OST3XgJaFWiSh9kJafzPK7ytJvUWMwIpnc5G1OUoP3U=","voting_power":10,"proposer_priority":10},{"address":"75FFEBE4D21B20E425A32026E97E3853F4680FD2","pub_key":"kx758TIgcgJlclsEUgXVYA/RVfAC6jfW1woiA/FfQ9E=","voting_power":10,"proposer_priority":10},{"address":"80506956036F1FC35DC82B7DCD0033AFD3AFC67C","pub_key":"Yqn9g4PteTr0jCNKI2gY5W0IeRa/OjrNjKgNjyqeSPo=","voting_power":10,"proposer_priority":10},{"address":"ACF5B6C87EF1823B49692FADE42C33A29A160E2D","pub_key":"gnX6QCqx0eei7Rb17o122MPhd6yYwlhFzQaReh47Fc0=","voting_power":10,"proposer_priority":10},{"address":"AFF83A98D6C337860622EF82167D1B4D92B41B11","pub_key":"ehEgoWgTExhN4OehRhwhyRUdXjgGm+M00PJGAFp5dfI=","voting_power":10,"proposer_priority":10},{"address":"D0EEF4C7075C1B06581230388FCFAFB1F839FB5E","pub_key":"9Nqr+uyl0g9p5KiisxgabhGDmeIkuhFoCI1vxVqrTH0=","voting_power":10,"proposer_priority":10},{"address":"D53759AFB4D5559F64370AA9F088362EDC83ED06","pub_key":"nQ49h9cUKeiLo6waIEcnIu6rNN75KUifxpClKL6lQWQ=","voting_power":10,"proposer_priority":10},{"address":"EF1F4CDB262078877CFA583B8DD1CEFB6D94A1A7","pub_key":"DaEyMg6mZmEKQd4Ep02liucF+CNp6tUZj53XpDjKj3U=","voting_power":10,"proposer_priority":10}],"proposer":{"address":"33D285936DEDAEB03B9497594130FEEADAFD566F","pub_key":"YdZcuYQh1nFOocmkbP+itvDFS9ZbpZkW4RFuyga9BX8=","voting_power":10,"proposer_priority":-80}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":2,"validator_address":"33D285936DEDAEB03B9497594130FEEADAFD566F","timestamp":"2023-12-13T23:46:39.361017-08:00","signature":"I+Q7fcyUQoD/fhiAQTkgQgf/8xjz87CKH3P6Jz2GJV3uNLOt/I6fe2srJPz93Z2ZyUcjWOn2r4Iqa0Y9DA+cCA=="},{"block_id_flag":2,"validator_address":"43C38459CE6D21F3CED9DA4F483E1E5779ACE736","timestamp":"2023-12-13T23:46:39.361048-08:00","signature":"se/U3oRycJxDYJvYdiR4R1+LD0BtLS9gHvvyX/VKSrxX1DOt6v32tfAUTXKVyRdXaRbRYSIcLFC6AUZroiZ3AA=="},{"block_id_flag":2,"validator_address":"75FFEBE4D21B20E425A32026E97E3853F4680FD2","timestamp":"2023-12-13T23:46:39.361077-08:00","signature":"upe54LPbylNLiSfYwKnOisrQGPe7qrLl1ZVfKuTyKXeDlqEAbqqrgKIrPYyYtN3nncuZLlslbuITEsCRdPdFAQ=="},{"block_id_flag":2,"validator_address":"80506956036F1FC35DC82B7DCD0033AFD3AFC67C","timestamp":"2023-12-13T23:46:39.36111-08:00","signature":"aKoFMiSf7c/614C7gkSKeG2dE4SZOLQZGOJpxP3SfAcAYV2Oqh5RYx+kwyone3F3iyqQkxSQ9Owr5BWhLRblAQ=="},{"block_id_flag":2,"validator_address":"ACF5B6C87EF1823B49692FADE42C33A29A160E2D","timestamp":"2023-12-13T23:46:39.361143-08:00","signature":"ZyI3Ihd2yd83BJuJz8tbWHnchRXwskK0ZvAjP2VtV7cZwvoXCva3dNUgViXhkXKw5qiDUtuz3l3alYxBr3wsAQ=="},{"block_id_flag":2,"validator_address":"AFF83A98D6C337860622EF82167D1B4D92B41B11","timestamp":"2023-12-13T23:46:39.361172-08:00","signature":"qlmw4V10/llMpZfAUHWoDiMPP37nw7Mtep8HYkc6XsHUpjZI9gR9wLZD1WD40DzRIsPAlZgVpuHO0rhPg8P7DA=="},{"block_id_flag":3,"validator_address":"D0EEF4C7075C1B06581230388FCFAFB1F839FB5E","timestamp":"2023-12-13T23:46:39.3612-08:00","signature":"b81BAIzhjxMQtBFK51l/dG83V1Jb6HcBjVdqPR1/CLWUTFUscawqgyRE0k4ZdA6nifgV6TDyQuk74w7Ny8sEDQ=="},{"block_id_flag":3,"validator_address":"D53759AFB4D5559F64370AA9F088362EDC83ED06","timestamp":"2023-12-13T23:46:39.361215-08:00","signature":"fis+L3g2x65pv2Vd+emeEm86jMkMKu7hN5PsINOUjfpuNrq92mOIjdubfYGnEMd0bMsjw5m2SKoXlAv1jAcvCA=="},{"block_id_flag":3,"validator_address":"EF1F4CDB262078877CFA583B8DD1CEFB6D94A1A7","timestamp":"2023-12-13T23:46:39.361229-08:00","signature":"IEOyWfNvHgMSxcOjqAeB0AjR1r2dA/KnKswfJ27dRfAtSpwNchnNl2RSRxnvyGK0bMSD4ff89asOiU5vEK0XDA=="}]}}`,

		`{"vs":{"validators":[{"address":"0FD1BBBF5A743354C8552749708FBBD704D8F3C2","pub_key":"9WGJsYAE3i1kzOod6LJxOrXMAZFn690OsCQGxBkkhbQ=","voting_power":10,"proposer_priority":-90},{"address":"54B0CB6F2218F2AA832363FCB24FF9E8F88BC04C","pub_key":"kJIhyZjvcLQJCaULAoNo41T6eadgj70WCTaWsF3Eqwc=","voting_power":10,"proposer_priority":10},{"address":"58DE9C3A9467FADE798F4D3F8738A80507DE565E","pub_key":"DQBsYxQVrsnIVM6BzdiOkkYGZuy3iLWnYv2IxvpRUdY=","voting_power":10,"proposer_priority":10},{"address":"8FC16FE8C406387E7977E6AAF7A7435E68C1CE0A","pub_key":"zjyU7ej08OIoH0F2A+bVy7kHnYAn9QHXrIMfOtWkWTc=","voting_power":10,"proposer_priority":10},{"address":"95D85693D6F91BA06036B7A93F76638B0F0402AA","pub_key":"dw87gRw+8VPjKUWxFdq8Bsd2clx/wsXkjw4TqGw+a3Y=","voting_power":10,"proposer_priority":10},{"address":"997F61F9057B91DB761E37D93E16163ABD9619A1","pub_key":"9Ufau1G0f8bDs/kCXlAj0WGbyrFg0VWU7wF9qD+mAK4=","voting_power":10,"proposer_priority":10},{"address":"A269D3AE75BE1DBDAFF251E75612D5BD46DD20B7","pub_key":"zXBOSnIw4vyvdUmwa4mqs3O7Wheb0UDbnCmGYhJ1Zko=","voting_power":10,"proposer_priority":10},{"address":"D25F5345AFA2DEC2C81EFCEBC531F4DD5BC05E2B","pub_key":"3Eq4Q66Jh/7wPXeqFftMbQD6j6Ay43sf9OsXuYDpt/U=","voting_power":10,"proposer_priority":10},{"address":"E895B44AE1114DB6A69487BD5AC86F8AFECC84DC","pub_key":"LhJBpN67AMYSg++nYvKEHe8xarR+XpplpEm5xz3BQrs=","voting_power":10,"proposer_priority":10},{"address":"EEE23E8C65C9867054ACFF6E70E3A1A3B45F3579","pub_key":"YupAFUfOG0NhkxUVEGA/Hx3eDBrVktUGatxvSQhwBZA=","voting_power":10,"proposer_priority":10}],"proposer":{"address":"0FD1BBBF5A743354C8552749708FBBD704D8F3C2","pub_key":"9WGJsYAE3i1kzOod6LJxOrXMAZFn690OsCQGxBkkhbQ=","voting_power":10,"proposer_priority":-90}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":1,"validator_address":"","timestamp":"0001-01-01T00:00:00Z","signature":null},{"block_id_flag":1,"validator_address":"","timestamp":"0001-01-01T00:00:00Z","signature":null},{"block_id_flag":1,"validator_address":"","timestamp":"0001-01-01T00:00:00Z","signature":null},{"block_id_flag":1,"validator_address":"","timestamp":"0001-01-01T00:00:00Z","signature":null},{"block_id_flag":1,"validator_address":"","timestamp":"0001-01-01T00:00:00Z","signature":null},{"block_id_flag":2,"validator_address":"997F61F9057B91DB761E37D93E16163ABD9619A1","timestamp":"2023-12-13T23:46:39.360192-08:00","signature":"oqNYxLnVTJbE4bCsUENEXCeLry8XIr7n42RD20LQtDHpWeErxxQMwPNROZT0hycbXqXtGB4PC6Rf7eOmhh9WAQ=="},{"block_id_flag":2,"validator_address":"A269D3AE75BE1DBDAFF251E75612D5BD46DD20B7","timestamp":"2023-12-13T23:46:39.360223-08:00","signature":"MI8MAsUu7BoMNTFLAoIPFBg+dueO5P7WobntPcmDxVklkDh0fBMQV8Qq0CAEd8/Fjywb6e0lTaecSoLoInuCCA=="},{"block_id_flag":2,"validator_address":"D25F5345AFA2DEC2C81EFCEBC531F4DD5BC05E2B","timestamp":"2023-12-13T23:46:39.360251-08:00","signature":"2i7dRMhCKeSObTPWqhNvIkpFnR1uPdNT+qlu2KskZ4fwGv6XDcmkvk+xD4Y32r/rgqOH8QvFw4WRN4kF+Me0BQ=="},{"block_id_flag":3,"validator_address":"E895B44AE1114DB6A69487BD5AC86F8AFECC84DC","timestamp":"2023-12-13T23:46:39.36028-08:00","signature":"24YbLcl2eg4ohRTIYKH8SFT1ipp4IfG3GYNjsG42XWdcv9ybqqvFA167A9KtMZ5jWEvifFy//saIAvQAM0W6Aw=="},{"block_id_flag":3,"validator_address":"EEE23E8C65C9867054ACFF6E70E3A1A3B45F3579","timestamp":"2023-12-13T23:46:39.360295-08:00","signature":"U+rzEx9fAaH/PbVos2b6RIXP6QvYxyr2v2uCB33Qo1peTXAIFlUULcpgn58gpcQxXPVkOUo16HYQPmsOAWEiBA=="}]}}`,

		`{"vs":{"validators":[{"address":"4E396EA189236134499BCAB69CD8B3CCFC120D48","pub_key":"RQKirc7gyTyP186uOKExay+JElyQzzqKsfWcwCPPO0M=","voting_power":10,"proposer_priority":-30},{"address":"6ED6958742DD191B8F6C99BE542BD44E1BDDAB7B","pub_key":"eHg6FC7ACHmW0RlMkZdb+yZvWamgyvO1w1iB4KJZOW8=","voting_power":10,"proposer_priority":10},{"address":"7021E94BC0E09BF2335D84546680BCA0DCE9296E","pub_key":"hKhuDASQ7SA6+Z+Xm6jsiDXj4QEyknTkOTbjd4KTwwc=","voting_power":10,"proposer_priority":10},{"address":"7150A344FB4F285A8FFC97D2F663E086127E435E","pub_key":"GuIyFeUEOnja/QL8uPj9lqyrC0hx7KHuvgq+n/E1mZs=","voting_power":10,"proposer_priority":10}],"proposer":{"address":"4E396EA189236134499BCAB69CD8B3CCFC120D48","pub_key":"RQKirc7gyTyP186uOKExay+JElyQzzqKsfWcwCPPO0M=","voting_power":10,"proposer_priority":-30}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":2,"validator_address":"4E396EA189236134499BCAB69CD8B3CCFC120D48","timestamp":"2023-12-13T23:46:39.359357-08:00","signature":"mESxgltpXVyH5E0IyHcwrILQ0bXCezDwHYJeUiZAI3/LPnCyjVmsJFipTMoydNajB8ygHqoKzZuS2jFEbqJqDg=="},{"block_id_flag":2,"validator_address":"6ED6958742DD191B8F6C99BE542BD44E1BDDAB7B","timestamp":"2023-12-13T23:46:39.359388-08:00","signature":"i6fXVWgx0q9VHplpfAkYuc95wiOaM4GWtJEgrR8tC9haa+WBKbrowID8bSpN5vbLUHroVxk7Yfyieh6RkhI/BQ=="},{"block_id_flag":2,"validator_address":"7021E94BC0E09BF2335D84546680BCA0DCE9296E","timestamp":"2023-12-13T23:46:39.359416-08:00","signature":"IyjIv9040Ox1hEDOq0r3KyQeTL6bkMU3qhX6IMN566gTX0QvFzQdmEbVHvuDXgrZxuF8JAbC9OTOKtPTsHVSAw=="}]}}`,

		`{"vs":{"validators":[{"address":"92BE9A0DF29959E3767FBD4696B7EAD17BAEE458","pub_key":"hY1kti4UlV/jlP/jDCuRwGA55w4wuIMp5xuXKMhBiZk=","voting_power":10,"proposer_priority":0}],"proposer":{"address":"92BE9A0DF29959E3767FBD4696B7EAD17BAEE458","pub_key":"hY1kti4UlV/jlP/jDCuRwGA55w4wuIMp5xuXKMhBiZk=","voting_power":10,"proposer_priority":0}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":1,"validator_address":"","timestamp":"0001-01-01T00:00:00Z","signature":null}]}}`,

		`{"vs":{"validators":[{"address":"8476D08A89B0F550C34D0201585751872A713449","pub_key":"RJdl97K+9ijlz6jzA6fmsHaAFVUZ58LCg7lkXV90Hy4=","voting_power":10,"proposer_priority":-10},{"address":"D254C658ABA946640DCE1EBAEBB0CC30307D949A","pub_key":"6xYx/cgzXe4FJAAeijxO2rLDh96lwSE1qjEuhrQVfFY=","voting_power":10,"proposer_priority":10}],"proposer":{"address":"8476D08A89B0F550C34D0201585751872A713449","pub_key":"RJdl97K+9ijlz6jzA6fmsHaAFVUZ58LCg7lkXV90Hy4=","voting_power":10,"proposer_priority":-10}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"D0B985398C210FDA13DB56D3C13C75A12C63B0C69A0615BA91D53D53A8A1749A","parts":{"total":123,"hash":"CABD71557A960196995DC619F5D2A2A61999AD1E459747E70EAF3D9B608A0A77"}},"signatures":[{"block_id_flag":2,"validator_address":"8476D08A89B0F550C34D0201585751872A713449","timestamp":"2023-12-13T23:46:39.358647-08:00","signature":"58tf0rudf6kUlzdzwo4h13whhTOhWGkEeSzbUzfURLkafJiCOrWQLjPy5ek8xmxpD6Z+sP1GBuuI45FsmL6ADg=="},{"block_id_flag":2,"validator_address":"D254C658ABA946640DCE1EBAEBB0CC30307D949A","timestamp":"2023-12-13T23:46:39.358678-08:00","signature":"3hSlRMJhpDDomqOzpfwv/Spje7LuLi4Bp5P7zNAp5mW6zRuZjd+QEpXK8NU+7OnUvBeY8LlsUXL2x+cbLZdxBg=="}]}}`,

		`{"vs":{"validators":[{"address":"7EB9B97A6EEEBCF9872BAEE35655688AF13EBA8D","pub_key":"JGYJ65O6QoHkLKB7VabpmjSzzHmFDYVtsy6UYtDJmrc=","voting_power":10,"proposer_priority":0}],"proposer":{"address":"7EB9B97A6EEEBCF9872BAEE35655688AF13EBA8D","pub_key":"JGYJ65O6QoHkLKB7VabpmjSzzHmFDYVtsy6UYtDJmrc=","voting_power":10,"proposer_priority":0}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":2,"validator_address":"7EB9B97A6EEEBCF9872BAEE35655688AF13EBA8D","timestamp":"2023-12-13T23:46:39.359736-08:00","signature":"jcQNb4iDsVEv4ipw7hnmF/gX7fxAPEZWgV+Jm3b8F0Bgyug6VS/VMboTg5ol1G6F234eZw0UtVpK70/NPdhnBA=="},{"block_id_flag":2,"validator_address":"7EB9B97A6EEEBCF9872BAEE35655688AF13EBA8D","timestamp":"2023-12-13T23:46:39.359766-08:00","signature":"8oSakg6QC65l4GVs8ANDnCQYaDPN+GshppiImGq6quyjKLcHxH+dIb7ufJqnIv2Lu5lyMkykqqZl86S1UivNCw=="}]}}`,

		`{"vs":{"validators":[{"address":"2F3C161479A2F0C2ACFA8AEB9AA7F7817BC2E306","pub_key":"sjq/+A0jHI/CuN+euerK6kvi0CuMyEp/Rczldzm0pAo=","voting_power":10,"proposer_priority":0}],"proposer":{"address":"2F3C161479A2F0C2ACFA8AEB9AA7F7817BC2E306","pub_key":"sjq/+A0jHI/CuN+euerK6kvi0CuMyEp/Rczldzm0pAo=","voting_power":10,"proposer_priority":0}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":99,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":2,"validator_address":"2F3C161479A2F0C2ACFA8AEB9AA7F7817BC2E306","timestamp":"2023-12-13T23:46:39.358995-08:00","signature":"jJTxadmjn/xPvhNcV1uck/27ShLCcfjfINW7z7lXRickxOuzAyMmIuNrtwqQQT/zO+2Ct83HHT3CvFCDhsPJCw=="}]}}`,

		`{"vs":{"validators":[{"address":"190D47129FA7CEEEAF9C86E9A16F035A5A8C8DD3","pub_key":"+RbqB4rlhEvs8RFLBNYREQAa7OAgfzrDaqtOgEhNzGo=","voting_power":10,"proposer_priority":-20},{"address":"2B67EED4F056AA5F41778EFB38F52A28A31ECA92","pub_key":"lzvimsA/6Vx6gHhC+hl5d2nIaBgxG81EYwwQWU1mii4=","voting_power":10,"proposer_priority":10},{"address":"5C91C1D92E9FE1B0EA960040F5681B93DAD355CD","pub_key":"vrb7n5IVTCEmaaC3nuV9Wvc2nxNShu443tmVd5zvHAI=","voting_power":10,"proposer_priority":10}],"proposer":{"address":"190D47129FA7CEEEAF9C86E9A16F035A5A8C8DD3","pub_key":"+RbqB4rlhEvs8RFLBNYREQAa7OAgfzrDaqtOgEhNzGo=","voting_power":10,"proposer_priority":-20}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":2,"validator_address":"190D47129FA7CEEEAF9C86E9A16F035A5A8C8DD3","timestamp":"2023-12-13T23:46:39.356506-08:00","signature":"tJgvqup1JSRVqxEjBq5myVadX5Rvbk+Py512Wo3ga6RdZHwdPJIYQAh6IELDFfVruxsA/dB6VE5UfyzJY6WBBA=="},{"block_id_flag":2,"validator_address":"2B67EED4F056AA5F41778EFB38F52A28A31ECA92","timestamp":"2023-12-13T23:46:39.356559-08:00","signature":"D+/uefNjpvROmzfRBLTSmfG7tNC3U7dm7zPKjV7yMfxzIzzrEx91ouG1z2+frjhOb34X+VXccpExJox882ICBQ=="},{"block_id_flag":2,"validator_address":"5C91C1D92E9FE1B0EA960040F5681B93DAD355CD","timestamp":"2023-12-13T23:46:39.356595-08:00","signature":"bgAH+f9cESzDFE089yCvnfMsrhLetip2mmD7dnZ0rl+4+WQxpTUdhf+0lCeiq5yYsNqja7NVqHiYGwHn3qOJAQ=="}]}}`,

		`{"vs":{"validators":[{"address":"6D9AA77C901B2634F8EA1A30900415059F329E07","pub_key":"XmCVoYe6gEhdeWyg1nKsXwDBBDIQ+knl8rSNwezK8v0=","voting_power":10,"proposer_priority":-10},{"address":"ECC31B465BB26FA2D5221B6CAEFAD46BA63328BC","pub_key":"poYJaTuI4qnVZU/vCHoNEvtxPzJtOFq1Ei/McWotxwo=","voting_power":10,"proposer_priority":10}],"proposer":{"address":"6D9AA77C901B2634F8EA1A30900415059F329E07","pub_key":"XmCVoYe6gEhdeWyg1nKsXwDBBDIQ+knl8rSNwezK8v0=","voting_power":10,"proposer_priority":-10}},"cid":"Lalande21185","bid":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"h":100,"comm":{"height":100,"round":0,"block_id":{"hash":"626C6F636B686173680000000000000000000000000000000000000000000000","parts":{"total":1000,"hash":"7061727473686173680000000000000000000000000000000000000000000000"}},"signatures":[{"block_id_flag":2,"validator_address":"6D9AA77C901B2634F8EA1A30900415059F329E07","timestamp":"2023-12-13T23:46:39.357866-08:00","signature":"wUF/7lgSQtVJTfXOl6bA5w//v/xjIBK1OXMd22xmBNP1T+KfoR1MQl6+MRQsdPyCNmcs8/iXdEuHIfNTlhd1Dw=="},{"block_id_flag":2,"validator_address":"ECC31B465BB26FA2D5221B6CAEFAD46BA63328BC","timestamp":"2023-12-13T23:46:39.357897-08:00","signature":"5Dm2jtBFWNWoPzqOWxe51L3VPxZpnarbinGQqszQpfZbBtow5VMYvCQEkY+HslnsV5rATkpmemVVjg6H9onWCg=="}]}}`,
	}

	// 1. Add seeds
	for _, seed := range seeds {
		f.Add(seed)
	}

	// 2. Run the fuzzers.
	f.Fuzz(func(t *testing.T, inputJSON string) {
		fsav := new(fuzzSave)
		if err := json.Unmarshal([]byte(inputJSON), fsav); err != nil {
			return
		}

		defer func() {
			r := recover()
			if r == nil {
				return
			}

			s, ok := r.(string)
			if !ok {
				panic(r)
			}

			if strings.Contains(s, "Unknown BlockIDFlag") {
				return
			}

			panic(r)
		}()

		valSet := fsav.ValSet
		_ = valSet.VerifyCommit(fsav.ChainID, fsav.BlockID, fsav.Height, fsav.Commit)
	})
}

type save struct {
	ChainID   string          `json:"ChainID"`
	ValSet    *ValidatorSet   `json:"ValSet"`
	ExtCommit *ExtendedCommit `json:"ExtCommit"`
}

func FuzzToExtendedVoteSet(f *testing.F) {
	if testing.Short() {
		f.Skip("-short enabled")
	}

	// Add seeds.
	matches, err := filepath.Glob(filepath.Join("testdata", "seeds", "fuzz-extcommit-*.json"))
	if err != nil {
		f.Fatal(err)
	}

	if len(matches) == 0 {
		f.Fatal("Could not load in seeds")
	}

	for _, matchFile := range matches {
		blob, err := os.ReadFile(matchFile)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(blob)
	}

	f.Fuzz(func(t *testing.T, inputJSON []byte) {
		ss := new(save)
		if err := json.Unmarshal(inputJSON, ss); err != nil {
			// Invalid data.
			return
		}

		chainID, valSet, extCommit := ss.ChainID, ss.ValSet, ss.ExtCommit
		if extCommit == nil {
			// Invalid data.
			return
		}
		if valSet == nil {
			// Invalid data.
			return
		}

		defer func() {
			r := recover()
			if r == nil {
				// There was no panic, just return.
				return
			}

			str := fmt.Sprintf("%v", r)
			switch {
			case strings.Contains(str, "height == 0, doesn't make sense"):
				return

			case strings.Contains(str, "Unknown BlockIDFlag:"):
				return

			case strings.Contains(str, "failed to validate vote reconstructed from LastCommit: expected ValidatorAddress"):
				return

			case strings.Contains(str, "failed to validate vote reconstructed from LastCommit"):
				return

			case strings.Contains(str, "failed to reconstruct vote set from extended commit"):
				return

			default: // This is an unhandled panic, re-throw it.
				panic(r)
			}
		}()

		_ = extCommit.ToExtendedVoteSet(chainID, valSet)
	})
}
