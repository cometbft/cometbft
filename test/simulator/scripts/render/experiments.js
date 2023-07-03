var chartDom = document.getElementById('main');
var myChart = echarts.init(chartDom);

$.getJSON(
    'experiments.json',
    function (data) {
	var symbolSize = 2.5;
	var option = {}
	
	option = {
	    grid3D: {},
	    xAxis3D: {
		name: '#nodes'
	    },
	    yAxis3D: {
		name: 'propagation rate (%)'
	    },
	    zAxis3D: {
		name: 'overhead'
	    },
	    dataset: {
		dimensions: [
		    'nodes',
		    'completion',
		    'propagation rate',
		    'overhead',
		],
		source: data
	    },
	    series: [
		{
		    type: 'scatter3D',
		    symbolSize: symbolSize,
		    grid3DIndex: 0,
		    encode: {
			x: 'nodes',
			y: 'propagation rate',
			z: 'overhead',
			tooltip: [0, 1, 2, 3, 4]
		    }
		}
	    ]
	};

	option && myChart.setOption(option);	
    }
);

