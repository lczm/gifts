import { useState } from "react";
import {
  TextField,
  Button,
  Stack,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Tooltip,
  IconButton,
} from "@mui/material";
import InfoIcon from "@mui/icons-material/Info";
import { lookupTeam, redeemGift } from "./api/gifts";
import "./App.css";

interface LookupResult {
  staff_pass_id: string;
  team_name: string;
  created_at: string;
}

interface RedemptionSuccess {
  team_name: string;
  redeemed_at: string;
}

interface RedemptionError {
  error: string;
}

function App() {
  // state for the entire application
  const [staffPassId, setStaffPassId] = useState("");
  const [lookupResult, setLookupResult] = useState<LookupResult | null>(null);
  const [redemptionResult, setRedemptionResult] = useState<
    RedemptionSuccess | RedemptionError | null
  >(null);

  // onclick handler to lookup
  const handleLookup = async () => {
    try {
      setRedemptionResult(null);
      const data = await lookupTeam(staffPassId);
      setLookupResult(data);
    } catch (error) {
      console.error("err : lookup failed:", error);
      setLookupResult(null);
    }
  };

  // onclick handler to redeem
  const handleRedeem = async () => {
    try {
      setLookupResult(null);
      const data = await redeemGift(staffPassId);
      setRedemptionResult(data);
    } catch (error) {
      console.error("err : redemption failed:", error);
    }
  };

  return (
    <Stack
      spacing={2}
      sx={{
        maxWidth: 800,
        margin: "auto",
        width: "100%",
        p: { xs: 1, sm: 2 },
      }}
    >
      <Typography variant="h4" component="h1" gutterBottom align="center">
        Gifts Redemption Counter
      </Typography>

      <Paper sx={{ p: { xs: 1, sm: 2 } }}>
        <Stack spacing={2} direction="column">
          <Stack direction="row" justifyContent="flex-end">
            <Tooltip
              title="The backend hosting this for convenience is using 'staff-id-to-team-mapping-long.csv', so any staff pass ID in that file can be used to test. But this can be ran locally as well. The API path is set in the .env file"
              placement="top"
            >
              <IconButton size="small">
                <InfoIcon />
              </IconButton>
            </Tooltip>
          </Stack>
          <TextField
            label="Staff Pass ID"
            value={staffPassId}
            onChange={(e) => setStaffPassId(e.target.value)}
            fullWidth
            size="medium"
          />
          <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
            <Button variant="contained" onClick={handleLookup} fullWidth>
              Lookup Team
            </Button>
            <Button variant="contained" onClick={handleRedeem} fullWidth>
              Redeem Gift
            </Button>
          </Stack>
        </Stack>
      </Paper>

      {lookupResult && !redemptionResult && (
        <TableContainer
          component={Paper}
          sx={{
            overflowX: "auto",
            "& .MuiTableCell-root": {
              px: { xs: 1, sm: 2 },
              py: { xs: 1, sm: 1.5 },
              whiteSpace: "nowrap",
            },
          }}
        >
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Staff Pass ID</TableCell>
                <TableCell>Team Name</TableCell>
                <TableCell>Created At</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <TableRow>
                <TableCell>{lookupResult.staff_pass_id}</TableCell>
                <TableCell>{lookupResult.team_name}</TableCell>
                <TableCell>
                  {new Date(lookupResult.created_at).toLocaleString()}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {redemptionResult && (
        <Paper
          sx={{
            p: { xs: 1, sm: 2 },
            bgcolor:
              "error" in redemptionResult ? "error.light" : "success.light",
            color:
              "error" in redemptionResult
                ? "error.contrastText"
                : "success.contrastText",
          }}
        >
          <Typography>
            {"error" in redemptionResult
              ? redemptionResult.error
              : `Gift redeemed successfully by team ${
                  redemptionResult.team_name
                } at ${new Date(
                  redemptionResult.redeemed_at
                ).toLocaleString()}`}
          </Typography>
        </Paper>
      )}
    </Stack>
  );
}

export default App;
