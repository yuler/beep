json.runs @runs do |run|
  json.partial! "api/v1/beepers/run", run: run
end
