!async function () {
    let data = await fetch("summary.json")
        .then((response) => response.json())
        .then(data => {
            return data;
        })
        .catch(error => {
            console.error(error);
        });

    var xData = [], yData = []
    var dupSum = 0
    for (i = 0; i < data.length; i++) {
        xData.push(data[i]['FileName1'] + "-" + data[i]['FileName2'])
        var dup = data[i]['DuplicateRate']
        yData.push(dup)
        dupSum += dup
    }

    console.log(xData, yData)

    // 基于准备好的dom，初始化echarts实例
    var myChart = echarts.init(document.getElementById('main'));

// 指定图表的配置项和数据
    var option = {
        title: {
            text: 'moss_plus statistics',
            subtext: `样本数量：${data.length} 总重复率：${dupSum}% 平均重复率：${dupSum / data.length}%`
        },
        tooltip: {},
        yAxis: {
            data: xData,
            inverse: true,
        },
        xAxis: {
            min: 0,
            max: 100
        },
        series: [
            {
                type: 'bar',
                data: yData,
                barWidth: '50%',
                itemStyle: {
                    normal: {
                        label: {
                            show: true, //开启显示
                            position: 'right', //在上方显示
                            textStyle: { //数值样式
                                color: 'black',
                                fontSize: 16
                            }
                        }
                    }
                },
            }
        ],
    };

// 使用刚指定的配置项和数据显示图表。
    myChart.setOption(option);
}();

