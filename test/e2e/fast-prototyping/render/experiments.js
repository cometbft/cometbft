var mainElt = document.getElementById('main');
var graphElt = document.getElementById('graph');
var mainChart = echarts.init(mainElt);
var graphChart = echarts.init(graphElt);

var headers = []
var headersMap = {
    "nodes" : "#nodes",
    "propagation_rate" : "propagation rate (%)",
    "sent" : "#sent",
    "seen" : "#seen (avg)",
    "completion" : "completion",
    "total_bandwidth" : "total bw. (KB)",
    "useful_bandwidth" : "useful bw. (KB)",
    "overhead" : "overhead",
    "redundancy" : "redundancy (avg)",
    "degree" : "degree (avg)",
    "bandwidth" : "bandwidth"
}
var dataset = []
var graphs = []

function plotExperiments(dimension) {
    var symbolSize = 2.5;
    var option = {}
    option = {
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
	    name: headersMap[dimension],

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
    var data = graphs[idx];

    var max = 0
    var min = Number.MAX_SAFE_INTEGER

    var option = {}
    var nodes = [];
    var edges = [];

    $.each(data, function (n, peers) {
        $.each(peers, function (m, bw) {
            max = Math.max(max, bw)
            min = Math.min(min, bw)
        });
    });

    var i = 0;

    $.each(data, function (n, peers) {
        r = Math.floor(Math.min(200,100 + Math.random() * 156));
        g = Math.floor(Math.min(200, 100 + Math.random() * 156));
        b = Math.floor(Math.min(200, 100 + Math.random() * 156));
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
        var j = 0;
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

    var datas = [];
    datas.push({
        nodes: nodes,
        edges: edges
    });

    option = {
	title: {
	    text: 'Bandwidth usage',
	    subtext: "(#nodes = " + dataset[idx]["nodes"] + ", propagation rate = " + dataset[idx]["propagation_rate"] + ")",
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

jQuery.get(
    'results.csv?'+Math.random(),
    function (csv) {
	var lines = csv.split("\n");
	var results = [];
	headers=lines[0].split(";");
	for(var i=1; i<lines.length; i++){
	    var obj = {};
	    var currentline=lines[i].split(";");
	    // skip empty lines
	    if (currentline.length > 1) {
		for(var j=0; j<headers.length-1; j++){
		    obj[headers[j]] = currentline[j];
		}
		dataset.push(JSON.parse(JSON.stringify(obj)));
		var graph = JSON.parse(currentline[j]);
		graphs.push(graph);
	    }
	}

	var dropdown = document.getElementsByClassName('dropdown-content')[0];
	for(var i=1; i<headers.length; i++){
	    if ( i != 0 && i != 1 && i != headers.length-1 ) {
		console.log(headers[i]);
		var child = document.createElement('a');
		child.setAttribute('onClick',"plotExperiments(\'"+headers[i]+"\')");
		child.innerHTML = headersMap[headers[i]];
		dropdown.appendChild(child);
	    }
	}

	plotExperiments('overhead')

	plotGraph(0)


    }
);
