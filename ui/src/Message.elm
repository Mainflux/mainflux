-- Copyright (c) 2019
-- Mainflux
--
-- SPDX-License-Identifier: Apache-2.0


module Message exposing (Model, Msg(..), initial, update, view)

import Bootstrap.Button as Button
import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Radio as Radio
import Bootstrap.Grid as Grid
import Bootstrap.Table as Table
import Bootstrap.Utilities.Spacing as Spacing
import Channel
import Error
import Helpers exposing (Globals)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import HttpMF exposing (path)
import List.Extra
import Thing
import Url.Builder as B


type alias Model =
    { message : String
    , thingkey : String
    , response : String
    , things : Thing.Model
    , channels : Channel.Model
    , thingid : String
    , checkedChannelsIds : List String
    }


initial : Model
initial =
    { message = ""
    , thingkey = ""
    , response = ""
    , things = Thing.initial
    , channels = Channel.initial
    , thingid = ""
    , checkedChannelsIds = []
    }


type Msg
    = SubmitMessage String
    | SendMessage
    | SentMessage (Result Http.Error String)
    | ThingMsg Thing.Msg
    | ChannelMsg Channel.Msg
    | SelectedThing String String Channel.Msg
    | CheckChannel String


update : Globals -> Msg -> Model -> ( Model, Cmd Msg )
update globals msg model =
    case msg of
        SubmitMessage message ->
            ( { model | message = message }, Cmd.none )

        SendMessage ->
            ( { model | message = "", thingkey = "", response = "", thingid = "" }
            , Cmd.batch
                (List.map
                    (\channelid -> send globals channelid model.thingkey model.message)
                    model.checkedChannelsIds
                )
            )

        SentMessage result ->
            case result of
                Ok statusCode ->
                    ( { model | response = statusCode }, Cmd.none )

                Err error ->
                    ( { model | response = Error.handle error }, Cmd.none )

        ThingMsg subMsg ->
            updateThing globals model subMsg

        ChannelMsg subMsg ->
            updateChannel globals model subMsg

        SelectedThing thingid thingkey channelMsg ->
            updateChannel globals { model | thingid = thingid, thingkey = thingkey, checkedChannelsIds = [] } (Channel.RetrieveChannelsForThing thingid)

        CheckChannel id ->
            ( { model | checkedChannelsIds = Helpers.checkEntity id model.checkedChannelsIds }, Cmd.none )


updateThing : Globals -> Model -> Thing.Msg -> ( Model, Cmd Msg )
updateThing globals model msg =
    let
        ( updatedThing, thingCmd ) =
            Thing.update globals msg model.things
    in
    ( { model | things = updatedThing }, Cmd.map ThingMsg thingCmd )


updateChannel : Globals -> Model -> Channel.Msg -> ( Model, Cmd Msg )
updateChannel globals model msg =
    let
        ( updatedChannel, channelCmd ) =
            Channel.update globals msg model.channels
    in
    ( { model | channels = updatedChannel }, Cmd.map ChannelMsg channelCmd )



-- VIEW


view : Model -> Html Msg
view model =
    Grid.container []
        [ Grid.row []
            [ Grid.col []
                [ Card.config
                    []
                    |> Card.headerH3 [] [ text "Things" ]
                    |> Card.block []
                        [ Block.custom
                            (Table.table
                                { options = [ Table.striped, Table.hover, Table.small ]
                                , thead =
                                    Table.simpleThead
                                        [ Table.th [] [ text "Name" ]
                                        , Table.th [] [ text "ID" ]
                                        ]
                                , tbody = Table.tbody [] <| genThingRows model.things.things.list
                                }
                            )
                        ]
                    |> Card.view
                , Html.map ThingMsg (Helpers.genPagination model.things.things.total Thing.SubmitPage)
                ]
            , Grid.col []
                [ Card.config
                    []
                    |> Card.headerH3 [] [ text "Channels" ]
                    |> Card.block []
                        [ Block.custom
                            (Table.table
                                { options = [ Table.striped, Table.hover, Table.small ]
                                , thead =
                                    Table.simpleThead
                                        [ Table.th [] [ text "Name" ]
                                        , Table.th [] [ text "ID" ]
                                        ]
                                , tbody = Table.tbody [] <| genChannelRows model.checkedChannelsIds model.channels.channels.list
                                }
                            )
                        ]
                    |> Card.view
                , Html.map ChannelMsg (Helpers.genPagination model.channels.channels.total Channel.SubmitPage)
                ]
            ]
        , Grid.row []
            [ Grid.col []
                [ Card.config []
                    |> Card.headerH3 [] [ text "Message" ]
                    |> Card.block []
                        [ Block.custom
                            (Form.form []
                                [ Form.group []
                                    [ Input.text [ Input.id "message", Input.onInput SubmitMessage ]
                                    ]
                                , Button.button [ Button.success, Button.attrs [ Spacing.ml1 ], Button.onClick SendMessage ] [ text "Send" ]
                                ]
                            )
                        ]
                    |> Card.view
                ]
            ]
        , Helpers.response model.response
        ]


genThingRows : List Thing.Thing -> List (Table.Row Msg)
genThingRows things =
    List.map
        (\thing ->
            Table.tr []
                [ Table.td [] [ label [] [ input [ type_ "radio", onClick (SelectedThing thing.id thing.key (Channel.RetrieveChannelsForThing thing.id)), name "things" ] [], text (Helpers.parseString thing.name) ] ]
                , Table.td [] [ text thing.id ]
                ]
        )
        things


genChannelRows : List String -> List Channel.Channel -> List (Table.Row Msg)
genChannelRows checkedChannelsIds channels =
    List.map
        (\channel ->
            Table.tr []
                [ Table.td [] [ input [ type_ "checkbox", onClick (CheckChannel channel.id), checked (Helpers.isChecked channel.id checkedChannelsIds) ] [], text (" " ++ Helpers.parseString channel.name) ]
                , Table.td [] [ text channel.id ]
                ]
        )
        channels



-- HTTP


send : Globals -> String -> String -> String -> Cmd Msg
send globals channelid thingkey message =
    HttpMF.request
        (B.crossOrigin globals.baseURL [ "http", path.channels, channelid, path.messages ] [])
        "POST"
        thingkey
        (Http.stringBody "application/json" message)
        SentMessage
