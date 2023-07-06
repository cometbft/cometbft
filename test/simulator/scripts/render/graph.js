var chartDom = document.getElementById('graph');
var myChart = echarts.init(chartDom);

$.getJSON(
    'graph.json',
    function (data) {
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
                if (i != j) {
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
            tooltip: {},
            animationDurationUpdate: 1500,
            animationEasingUpdate: 'quinticInOut',
            series: {
                type: 'graph',
                layout: 'circular',
                data: nodes,
                links: edges,
                roam: true,
                labelLayout: {
                    hideOverlap: true
                },
                emphasis: {
                    focus: 'self'
                },
            }
        }

        option && myChart.setOption(option);
    }

);
