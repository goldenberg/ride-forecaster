console.log("forecast.js loaded");

angular.module('forecasterApp', ['n3-line-chart', 'ngMap'])
    .controller('ForecastController', function($scope, $http) {
        $scope.route = "bofax_alpine11.gpx";
        $scope.startTime = new Date(2015, 0, 01, 08, 0, 0);
        $scope.velocity = 11;
        var forecaster = this;
        var measurementUnits = {
            "temperature": "° F",
            "windSpeed": " mph",
            "precipAccumulation": " in/hr",
            "heading": "°",
            "windAngle": "°",
        }
        $scope.options = {
            axes: {
                x: {
                    key: "x",
                    type: "date",
//                    min: new Date($scope.startTime.getTime()),
//                    max: new Date($scope.startTime.getTime() + 4 * 60 * 60 * 1000) // 4 hours later
//                    zoomable: true
                },
                y: {
                    key: "temperature",
                },
                y2: {
                    key: "precipAccumulation",
                }
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
                mode: "scrubber",
                formatter: function(x, y, series) {
                    var unit = measurementUnits[series.y];
                    return x.getHours() + ":" + x.getUTCMinutes() + "  " +
                        Math.round(y * 10) / 10. + unit;
                }
            }
        };

        $scope.data = [{
            x: $scope.startTime,
            temperature: 0,
            windSpeed: 0,
            precipAccumulation: 0,
            heading: 0,
            windAngle: 0,

        }];

        $scope.routePath = [];
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

                    var path = $scope.poly.getPath();

                    for (var i in resp) {
                        var currently = resp[i].forecast.currently;
                        $scope.data.push({
                            "x": new Date(resp[i].waypoint.time),
                            temperature: currently.temperature,
                            windSpeed: currently.windSpeed,
                            precipAccumulation: currently.precipAccumulation,
                            heading: resp[i].heading,
                            windAngle: resp[i].windAngle
                        });

                        var waypt = resp[i].waypoint.Point;
                        // var path = $scope.path //poly.getPath();
                        console.log(waypt[0]);
                        // path.push(google.maps.LatLng(waypt[0], waypt[1]));
                        $scope.routePath.push([waypt[0], waypt[1]]);
                    };
                    // $scope.poly.setPath(path);
                })
                .error(function(data, status, headers, config) {
                    $scope.status = "Got error status " + status + ": " + data;
                });
        };

        $scope.googleMapsUrl="http://maps.google.com/maps/api/js?v=3.20&key=AIzaSyA6rVO54LQO9Ln0qyQHv6Gh_Llo3xO_HVs"

        $scope.$on('mapInitialized', function(event, map) {
            // $scope.poly = new google.maps.Polyline({
            //   strokeColor: '#000000',
            //   strokeOpacity: 1.0,
            //   strokeWeight: 3,
            //   // path: []
            //   // path: [
            //     // {lat: 37.772, lng: -122.214},
            //     // {lat: 21.291, lng: -157.821}
            //   // ]
            // });
            // // $scope.poly = poly
            // $scope.poly.setMap(map);
            // console.log($scope.poly.getPath());
            console.log("initing map");

        })


    });