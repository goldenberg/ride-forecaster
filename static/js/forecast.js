console.log("forecast.js loaded");

angular.module('forecasterApp', ['n3-line-chart'])
    .controller('ForecastController', function($scope, $http) {
        var forecaster = this;
        $scope.options = {
            axes: {
                x: {
                    key: "x",
                    type: "date"
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
            },
            ]
        };

        $scope.data = [{
            x: 0,
            temperature: 0,
            windSpeed: 0

        }, {
            x: 1,
            temperature: 0.993,
            windSpeed: 3.894,
        }, {
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
        }];
        forecaster.submit = function() {
            var url = "http://localhost:8080/forecast";// + $("forecastParams").serialize();
            var params = {
                "route": $scope.route,
                "startTime": $scope.startTime,
                "velocity": $scope.velocity,
            };
            $scope.rawDataURL = url;
            $scope.status = "Fetching data from: " + url;
            $http.get(url, params).
                success(function(resp, status, headers, config) {
                    $scope.status = "Received resp " + resp;
                    var newData = []
                    for (var i in resp) {
                        var currently = resp[i].forecast.currently;

                        console.log("pushing " + currently);
                        newData.push({
                            "x": currently.time,
                            temperature: currently.temperature,
                            windSpeed: currently.windSpeed,
                            precipAccumulation: currently.precipAccumulation,
                            heading: resp[i].heading,
                            windAngle: resp[i].windAngle
                        });
                    };
                    $scope.data = newData;
                })
                .error(function(data, status, headers, config) {
                    // TODO
                });
        };
    });