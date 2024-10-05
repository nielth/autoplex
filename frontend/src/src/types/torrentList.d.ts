interface TorrentData {
  facets: Object;
  facetswoc: Object;
  numFound: number;
  orderBy: string;
  page: number;
  perPage: number;
  torrentList: [
    {
      addedTimestamp: string;
      categoryID: number;
      commentsDisabled: number;
      completed: number;
      download_multiplier: number;
      fid: string;
      filename: string;
      genres: string;
      igdbID: string;
      imdbID: string;
      leechers: number;
      name: string;
      new: boolean;
      numComments: number;
      rating: number;
      seeders: number;
      size: number;
      tags: [string];
      tvmazeID: string;
      uploader: string;
    },
  ];
  userTimeZone: string;
}
