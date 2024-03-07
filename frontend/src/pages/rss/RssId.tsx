import { Box, FormControlLabel, Tooltip, Typography } from "@mui/material";
import axios from "axios";
import { Fragment, useState } from "react";
import { useParams } from "react-router-dom";
import Switch from "@mui/material/Switch";

const json = {
  id: 82,
  url: "https://www.tvmaze.com/shows/82/game-of-thrones",
  name: "Game of Thrones",
  type: "Scripted",
  language: "English",
  genres: ["Drama", "Adventure", "Fantasy"],
  status: "Ended",
  runtime: 60,
  averageRuntime: 61,
  premiered: "2011-04-17",
  ended: "2019-05-19",
  officialSite: "http://www.hbo.com/game-of-thrones",
  schedule: { time: "21:00", days: ["Sunday"] },
  rating: { average: 8.9 },
  weight: 98,
  network: {
    id: 8,
    name: "HBO",
    country: {
      name: "United States",
      code: "US",
      timezone: "America/New_York",
    },
    officialSite: "https://www.hbo.com/",
  },
  webChannel: null,
  dvdCountry: null,
  externals: { tvrage: 24493, thetvdb: 121361, imdb: "tt0944947" },
  image: {
    medium:
      "https://static.tvmaze.com/uploads/images/medium_portrait/498/1245274.jpg",
    original:
      "https://static.tvmaze.com/uploads/images/original_untouched/498/1245274.jpg",
  },
  summary:
    "<p>Based on the bestselling book series <i>A Song of Ice and Fire</i> by George R.R. Martin, this sprawling new HBO drama is set in a world where summers span decades and winters can last a lifetime. From the scheming south and the savage eastern lands, to the frozen north and ancient Wall that protects the realm from the mysterious darkness beyond, the powerful families of the Seven Kingdoms are locked in a battle for the Iron Throne. This is a story of duplicity and treachery, nobility and honor, conquest and triumph. In the <b>Game of Thrones</b>, you either win or you die.</p>",
  updated: 1704792924,
  _links: {
    self: { href: "https://api.tvmaze.com/shows/82" },
    previousepisode: { href: "https://api.tvmaze.com/episodes/1623968" },
  },
};

export default function RssId() {
  const [info, setInfo] = useState<any>();
  const { rssId } = useParams();
  console.log(rssId);

  console.log(json);

  //   axios(`https://api.tvmaze.com/shows/${rssId}`).then((respInfo) => {
  //     setInfo(respInfo.data);
  //   });

  return (
    <>
      <Box display={"flex"} gap={3}>
        <Box component={"img"} height={300} zIndex={-999} src={json.image.original} />
        <Box display={"flex"} flexDirection={"column"}>
          <Typography textAlign={"left"} variant="h4">
            {json.name}
          </Typography>
          <Box pt={2} textAlign={"left"} alignItems={"flex-start"}>
            <Typography>Premiere: {json.premiered}</Typography>
            <Typography>End: {json.ended ? json.ended : "Running"}</Typography>
          </Box>
          <Box pt={2} display={"flex"} textAlign={"left"}>
            <Tooltip
              title={
                <Fragment>
                  <Box textAlign={"left"}>
                    <Typography>Coming episodes</Typography>
                    <Typography>(not complete series)</Typography>
                  </Box>
                </Fragment>
              }
              arrow
            >
              <FormControlLabel
                sx={{ m: 0 }}
                control={<Switch />}
                label={<Typography>Auto Download:</Typography>}
                labelPlacement="start"
              />
            </Tooltip>
          </Box>
        </Box>
      </Box>
    </>
  );
}
