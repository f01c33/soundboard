#!/bin/bash

ffmpeg -i $1 -af silenceremove=start_periods=1:stop_periods=1:detection=peak fast_$1
mv fast_$1 $1