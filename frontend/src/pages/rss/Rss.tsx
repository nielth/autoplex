import { Box, Button, Grid, Paper, TextField, Typography } from "@mui/material";
import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

const json = [
  {
    score: 1.4282309,
    show: {
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
    },
  },
  {
    score: 1.0242988,
    show: {
      id: 42281,
      url: "https://www.tvmaze.com/shows/42281/game-of-thrones-inside-the-episode",
      name: "Game of Thrones: Inside the Episode",
      type: "Documentary",
      language: "English",
      genres: [],
      status: "Ended",
      runtime: 8,
      averageRuntime: 8,
      premiered: "2019-04-14",
      ended: "2019-05-12",
      officialSite: "https://www.youtube.com/channel/UCQzdMyuz0Lf4zo4uGcEujFw",
      schedule: { time: "18:00", days: ["Sunday"] },
      rating: { average: null },
      weight: 65,
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
      webChannel: {
        id: 21,
        name: "YouTube",
        country: null,
        officialSite: "https://www.youtube.com",
      },
      dvdCountry: null,
      externals: { tvrage: null, thetvdb: null, imdb: null },
      image: null,
      summary: "<p>Behind the scenes of Game of Thrones.</p>",
      updated: 1561639292,
      _links: {
        self: { href: "https://api.tvmaze.com/shows/42281" },
        previousepisode: { href: "https://api.tvmaze.com/episodes/1660268" },
      },
    },
  },
  {
    score: 0.91415393,
    show: {
      id: 22602,
      url: "https://www.tvmaze.com/shows/22602/hip-hop-tribe-2-game-of-thrones",
      name: "Hip Hop Tribe 2 : Game of Thrones",
      type: "Variety",
      language: "Korean",
      genres: ["Music"],
      status: "Ended",
      runtime: 90,
      averageRuntime: 90,
      premiered: "2016-10-18",
      ended: "2017-01-24",
      officialSite: null,
      schedule: { time: "", days: [] },
      rating: { average: null },
      weight: 28,
      network: {
        id: 268,
        name: "jTBC",
        country: {
          name: "Korea, Republic of",
          code: "KR",
          timezone: "Asia/Seoul",
        },
        officialSite: null,
      },
      webChannel: null,
      dvdCountry: null,
      externals: { tvrage: null, thetvdb: null, imdb: null },
      image: {
        medium:
          "https://static.tvmaze.com/uploads/images/medium_portrait/119/298108.jpg",
        original:
          "https://static.tvmaze.com/uploads/images/original_untouched/119/298108.jpg",
      },
      summary:
        "<p>A hip hop and rap competition program where older generation contestants are teamed up with professional hip hop music producers.</p>",
      updated: 1561639241,
      _links: {
        self: { href: "https://api.tvmaze.com/shows/22602" },
        previousepisode: { href: "https://api.tvmaze.com/episodes/1124799" },
      },
    },
  },
];

export default function Rss() {
  const [search, setSearch] = useState("");
  const [resp, setResp] = useState(json);
  const handleChange = (event: any) => {};
  const navigate = useNavigate();

  const handleSearch = (event: any) => {
    if (event.key === "Enter") {
      setSearch(event.target.value);
    }
  };

  useEffect(() => {
    if (search !== "") {
      axios
        .get(`https://api.tvmaze.com/search/shows?q=${search}`)
        .then((tvmaze) => {});
    }
  });

  return (
    <>
      <Box textAlign={"center"}>
        <TextField
          label="Auto Download"
          variant="outlined"
          onChange={handleChange}
          onKeyDown={handleSearch}
          color="primary"
          sx={{ width: "50%" }}
        />
      </Box>
      <Box display={"flex"} flexDirection={"column"} gap={1} pt={5}>
        {json.map((row) => (
          <Box>
            <Box component={Button} display={"flex"} flexDirection={"column"} onClick={(() => {
                navigate(`/rss/${row.show.id}`)
            })}>
              <Typography>{row.show.name}</Typography>
            </Box>
          </Box>
        ))}
      </Box>
    </>
  );
}
