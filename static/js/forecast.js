console.log("forecast.js loaded");

angular.module('forecasterApp', ['n3-line-chart'])
    .controller('ForecastController', function($scope, $http) {
        var forecaster = this;
        $scope.options = {
            axes: {
                x: {
                    key: "x"
                },
            },
            series: [{
                y: "temperature",
                label: "Temperature",
                color: "#1f77b4"
            }, {
                y: "windSpeed",
                label: "Wind Speed",
                color: "#ff7f0e",
            }]
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
        $scope.message = "foo";
        //                        console.log("data" + $scope.data + " with response size" + resp.size);
        //                for (var i in resp) {
        //                    var currently = resp[i].forecast.currently;
        //                    console.log("pushing " + currently);
        //                    $scope.data.push({
        //                        "x": currently.time,
        //                        temperature: currently.temperature,
        //                        windSpeed: currently.windSpeed,
        //                        // precipAccumulation: x.currently.precipAccumulation,
        //                        // time: x.waypoint.time,
        //                        // heading: x.waypoint.heading,
        //                        // windAngle: x.waypoint.windAngle
        //                    });
        //                };
        //                        console.log("data is now " + $scope.data[0])
        //      }
        forecaster.submit = function() {
            $scope.data = [{
                x: 0,
                temperature: 0,
                windSpeed: 0,
            }, {
                x: 1,
                temperature: 10,
                windSpeed: 15,
            }];
            console.log("submit button clicked");
            var url = "http://localhost:8080/forecast?" + $("forecastParams").serialize();
            console.log("url: " + url);
            x = "bar";
            $scope.message = "bar";
            //            $http.get("/static").success(updateData);
        };
        $scope.message = "baz";

    });