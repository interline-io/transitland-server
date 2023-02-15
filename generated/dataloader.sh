#!/bin/sh

REPO="github.com/interline-io/transitland-server/model"

rm dataloader/*_gen.go

arrayName=( Feed Agency Calendar Route Stop Level Shape FeedVersion FeedState FeedVersionGtfsImport RouteHeadway CensusTable Trip Operator StopExternalReference )
arrayWhereName=( Feed FeedVersion FeedVersionFileInfo FeedVersionServiceLevel FeedFetch Agency Route StopTime Trip Frequency RouteStop RouteHeadway RouteGeometry Stop AgencyPlace Operator CensusGeography CensusValue Pathway FeedInfo CalendarDate RouteStopPattern StopObservation )

cwd=${PWD}
cd ${PWD}/dataloader

for i in "${arrayName[@]}"
do
    : 
   go run github.com/vektah/dataloaden ${i}Loader int "*${REPO}.${i}"
done

for i in "${arrayWhereName[@]}"
do
    :
   go run github.com/vektah/dataloaden ${i}WhereLoader "${REPO}.${i}Param" "[]*${REPO}.${i}"
done

cd ${cwd}