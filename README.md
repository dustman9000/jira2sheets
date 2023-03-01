# jira2sheets

A tool to query a JIRA filter & upload to Google Sheets.

## Building

```
make
```

## Running

```
export JIRA2SHEETS_GOOGLE_CREDENTIALS_JSON=`cat ./my-google-credentials.json`
export JIRA2SHEETS_JIRA_PAT=`cat ./my-jira-personal-access-token`
./jira2sheets import -v -c samples/my-config.yml -v
```
