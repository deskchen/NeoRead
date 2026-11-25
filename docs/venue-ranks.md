# Offline venue rankings

The journal indicators include two offline rankings. The build embeds only the
published venue names, abbreviations, and categories needed to match the selected
grades. CCF, CORE, and THCPL are selected by default. CCF uses the configured
easyScholar account. A confirmed lookup without a CCF rank displays
`Non-CCF`, as does an arXiv-only preprint. Failed lookups do not establish
whether a venue is indexed.

- **CORE A\*** uses the A* conference records in the [official ICORE 2026 export](https://portal.core.edu.au/conf-ranks/?by=all&source=ICORE2026&do=Export).
- **THCPL A** uses the A conference and journal records in the [Tsinghua Computer Science Preferred List](https://numbda.cs.tsinghua.edu.cn/~yuwj/TH-CPL.pdf) approved August 13, 2019.

The badges are ordered CCF, CORE, and THCPL. `Non-CORE` and `Non-THCPL`
indicate that the venue did not match the embedded A* or A list, respectively.
The badge tooltip shows the ranking edition when available. The year shown
beside a conference abbreviation comes from the proceedings name or the paper's
publication date and is displayed as a two-digit suffix, for example `ASPLOS'25`.
Venue matching uses published names and unambiguous aliases
from the conference directory.
