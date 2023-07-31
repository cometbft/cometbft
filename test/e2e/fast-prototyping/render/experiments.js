const mainElt = document.getElementById('main');
const graphElt = document.getElementById('graph');
const mainChart = echarts.init(mainElt);
const graphChart = echarts.init(graphElt);

const headersMap = {
    "nodes" : "#nodes",
    "propagation_rate" : "propagation rate (%)",
    "submitted" : "#submitted (total)",
    "added" : "#added as valid (avg)",
    "sent" : "#sent (avg)",
    "completion" : "completion",
    "total_mempool_bandwidth" : "total mempool bw. (KB)",
    "useful_mempool_bandwidth" : "useful mempool bw. (KB)",
    "overhead" : "overhead",
    "redundancy" : "redundancy", // average
    "degree" : "degree (avg)", // average out+in degree
    "cpu_load" : "CPU load (avg)", // average cpu load (s)
    "latency" : "latency ", // average in #blocks (or seconds if physical)
    "bandwidth" : "bandwidth"
}
let headers = []
let dataset = []
let graphs = []

let metric = "overhead"
let index = 0

function plotExperiments(dimension) {
    metric = dimension
    let symbolSize = 2.5;
    let option = {
        title: {
            text: 'Experiments',
            subtext: 'e2e - fast prototyping',
            left: 'center'
        },
        grid3D: {},
        xAxis3D: {
            name: headersMap['nodes']
        },
        yAxis3D: {
            name: headersMap['propagation_rate']
        },
        zAxis3D: {
            name: headersMap[metric],

        },
        dataset: {
            dimensions: headers,
            source: dataset
        },
        series: [
            {
                name: 'experiments',
                type: 'scatter3D',
                symbolSize: 10,
                encode: {
                    x: 'nodes',
                    y: 'propagation_rate',
                    z: dimension,
                    tooltip: [0, 1, 2, 3, 4]
                }
            }
        ]
    };

    option && mainChart.setOption(option);

    mainChart.on('click', {seriesName: 'experiments'}, function (params) {
        graphChart.clear();
        plotGraph(params.dataIndex);
    });

}

function plotGraph(idx) {
    index = idx

    let data = graphs[index];

    let max = 0
    let min = Number.MAX_SAFE_INTEGER

    let option = {}
    let nodes = [];
    let edges = [];

    $.each(data, function (n, peers) {
        $.each(peers, function (m, bw) {
            max = Math.max(max, bw)
            min = Math.min(min, bw)
        });
    });

    let i = 0;

    $.each(data, function (n, peers) {
        let r = Math.floor(Math.min(200, 100 + Math.random() * 156));
        let g = Math.floor(Math.min(200, 100 + Math.random() * 156));
        let b = Math.floor(Math.min(200, 100 + Math.random() * 156));
        nodes.push(
            {
                name: n,
                label: {
                    show: true,
                    position: 'inside',
                    fontSize: 20
                },
                symbolSize: 50,
                itemStyle: {
                    color: 'rgb('+r+','+g+','+b+')'
                }
            }
        );
        let j = 0;
        $.each(peers, function (m, bw) {
            if (i != j && bw != 0) {
                edges.push(
                    {
                        source: n,
                        target: j,
                        lineStyle: {
                            width: 1,
                            curveness: 0.05,
                            opacity: 0.05 + ((bw-min)/(max-min)*0.3),
                            color: '#444'
                        },
                        label: {
                            show: true,
                            position: 'middle',
                            fontSize: '20',
                            opacity: 1,
                            color: '#000',
                            formatter: function () {
                                return Math.floor(bw / 100) + " KB";
                            }
                        }
                    }
                );
            }
            j++
        });
        i++;
    });

    let datas = [];
    datas.push({
        nodes: nodes,
        edges: edges
    });

    option = {
	title: {
	    text: 'Bandwidth usage',
	    subtext: "(#nodes = " + dataset[index]["nodes"] + ", propagation rate = " + dataset[index]["propagation_rate"] + ")",
	    left: 'center'
	},
        tooltip: {},
        animationDurationUpdate: 1500,
        animationEasingUpdate: 'quinticInOut',
        series: {
	    name: 'graph',
            type: 'graph',
            layout: 'circular',
            data: nodes,
            links: edges,
            roam: true,
            edgeSymbol: ['circle', 'arrow'],
            edgeSymbolSize: [4, 10],
            labelLayout: {
                hideOverlap: true
            },
            emphasis: {
                focus: 'self'
            },
        }
    }

    option && graphChart.setOption(option);

}

function displayResults(file) {
    jQuery.get(
        file + '?' + Math.random(), // to invalid cache
        function parse(csv) {
            dataset = []
            graphs = []

            const lines = csv.split("\n");
            headers = lines[0].split(";");
            for (i = 1; i < lines.length; i++) {
                const obj = {};
                const currentLine = lines[i].split(";");
                // skip empty lines
                if (currentLine.length > 1) {
                    for (j = 0; j < headers.length - 1; j++) {
                        obj[headers[j]] = currentLine[j];
                    }
                    dataset.push(JSON.parse(JSON.stringify(obj)));
                    const graph = JSON.parse(currentLine[j]);
                    graphs.push(graph);
                }
            }

            const dropdown = document.getElementsByClassName('dropdown-content')[0];
            dropdown.textContent = '';
            for (i = 1; i < headers.length; i++) {
                if (i != 0 && i != 1 && i != headers.length - 1) {
                    var child = document.createElement('a');
                    child.setAttribute(
                        'onClick',
                        "plotExperiments(\'" + headers[i] + "\')");
                    child.innerHTML = headersMap[headers[i]];
                    dropdown.appendChild(child);
                }
            }

            plotExperiments(metric, headers)

            plotGraph(index)

        }
    );
}

jQuery.get(
    'results.csv?'+Math.random(),
    function load(csv) {
        const lines = csv.split("\n");
        const results = [];
        const headers = lines[0].split(";");
        for (i = 1; i < lines.length; i++) {
            const obj = {};
            const currentLine = lines[i].split(";");
            if (currentLine.length > 1) {
                for (j = 0; j < currentLine.length; j++) {
                    obj[headers[j]] = currentLine[j];
                }
                results.push(obj)
            }
        }

        const row = document.getElementsByClassName('results-content')[0];
        const resultHeader = document.createElement('th');
        resultHeader.setAttribute('class', 'results-header');
        resultHeader.innerHTML = "Navigate the results:";
        row.appendChild(resultHeader);

        for (i = 0; i < results.length; i++) {
            const child = document.createElement('td')
            const button = document.createElement('button');
            button.setAttribute('onClick', "displayResults(\'" + results[i]["location"] + "\')");
            button.setAttribute('title', "\'" + results[i]["description"] + "\'");
            button.setAttribute('class', 'results-button');
            button.innerHTML = results[i]["name"];
            child.appendChild(button);
            row.append(child);
        }

        displayResults(results[0]['location']);

    }
);
