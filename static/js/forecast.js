console.log("forecast.js loaded");

angular.module('forecasterApp', ['n3-line-chart'])
    .controller('ForecastController', function($scope, $http) {
        $scope.route = "bofax_alpine11.gpx";
        $scope.startTime = new Date();
        $scope.velocity = 11;
        var forecaster = this;
        $scope.options = {
            axes: {
                x: {
                    key: "x",
                    type: "date",
                    min: $scope.startTime.getTime(),
                    max: $scope.startTime.getTime() + 12 * 60 * 60 * 1000 // 12 hours later
                },
            },
            series: [{
                y: "temperature",
                label: "Temperature",
            }, {
                y: "windSpeed",
                label: "Wind Speed",
            },  {
                y: "precipAccumulation",
                label: "Precipitation Accumulation",
                type: "column"
            },  {
                y: "heading",
                label: "Heading"
            },  {
                y: "windAngle",
                label: "Wind Angle",
            }
            ],
            tooltip: {
                mode: "axes",
//                formatter: function(x, y, series) {
//                    return moment(x).fromNow() + " : " + y;
//                }
            }
        };

//        $scope.data = [];
        $scope.data = [{
            x: $scope.startTime,
            temperature: 0,
            windSpeed: 0,
            precipAccumulation: 0,
            heading: 0,
            windAngle: 0,

        }];  /*{
            x: $scope.startTime + 61 * 60 * 1000,
            temperature: 0.993,
            windSpeed: 3.894,
            precipAccumulation: 0,
            heading: 0,
            windAngle: 0,
        }, ]; /*{
            x: 2,
            temperature: 1.947,
            windSpeed: 7.174,
        }, {
            x: 3,
            temperature: 2.823,
            windSpeed: 9.32,
        }, {
            x: 4,
            temperature: 3.587,
            windSpeed: 9.996,
        }, {
            x: 5,
            temperature: 4.207,
            windSpeed: 9.093,
        }, {
            x: 6,
            temperature: 4.66,
            windSpeed: 6.755,
        }]; */
        forecaster.submit = function() {
            var url = "http://localhost:8080/forecast";
            var params = {
                "route": $scope.route,
                "startTime": $scope.startTime,
                "velocity": $scope.velocity,
            };
            $scope.rawDataURL = url;
            $scope.status = "Fetching data from: " + url;
            $http.get(url, {"params": params}).
                success(function(resp, status, headers, config) {
                    $scope.status = "Received resp " + resp;
                    $scope.data = [];
                    for (var i in resp) {
                        var currently = resp[i].forecast.currently;
                        $scope.data.push({
                            "x": currently.time,
                            temperature: currently.temperature,
                            windSpeed: currently.windSpeed,
                            precipAccumulation: currently.precipAccumulation,
                            heading: resp[i].heading,
                            windAngle: resp[i].windAngle
                        });
                    };
                })
                .error(function(data, status, headers, config) {
                    // TODO
                });
        };
    });