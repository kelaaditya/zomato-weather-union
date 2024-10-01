#
# program takes in: 
#   temperature             [celcius]
#   humidity (relative)     [percentage]
#   pressure                [hPa]
# and returns
#   wet_bulb_temperature    [celcius]

import metpy
from metpy.units import units
from metpy.calc import dewpoint_from_relative_humidity
from metpy.calc import wet_bulb_temperature
import argparse
import json
import sys

# initialize parser
parser = argparse.ArgumentParser()
parser.add_argument('--temperature', type=float, required=True)  # in celcius
parser.add_argument('--humidity', type=float, required=True)  # in percentage
parser.add_argument('--pressure', type=float, required=True)  # in hPa
# parse the arguments
args = parser.parse_args()


# calculate dewpoint
# arguments:
#   temperature
#   humidity (relative)
temperature_estimate_dew_point = dewpoint_from_relative_humidity(
    args.temperature * units.degC,
    args.humidity * units.percent
)

# calculate wet bulb estimate
# arguments:
#   pressure
#   temperature
#   dewpoint
temperature_estimate_wet_bulb = wet_bulb_temperature(
    args.pressure * units.hPa,
    args.temperature * units.degC,
    temperature_estimate_dew_point.magnitude * units.degC
)

# build json
return_data = {}
return_data["temperature_dew_point"] = temperature_estimate_dew_point.magnitude
return_data["temperature_wet_bulb"] = temperature_estimate_wet_bulb.magnitude
# write to stdout so that
# golang can capture it
json.dump(return_data, sys.stdout)