# Gifts

Gift Redemption System. The system can either look up a staff, or let the staff, as a representative, redeem a gift from the system.

Setup

```bash
git clone https://github.com/lczm/gifts
```

`/api`, golang backend

```
cd api
go run . -csv=staff-id-to-team-mapping-long.csv
```

- To run the tests, `go test`

`/web`, frontend

```
cd web
npm install
npm run dev
```

For convenience, the frontend has been [deployed on github pages](https://lczm.github.io/gifts/). It's been set to point to my backend, that's using `staff-id-to-team-mapping-long.csv`.
