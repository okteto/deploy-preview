#!/usr/bin/env ruby
require "octokit"
require "json"

if ENV["GITHUB_EVENT_NAME"] != "pull_request"
    puts "This action only supports pull_request events."
    exit(1)
end

if !message = ARGV[1]
    puts "Missing GITHUB_TOKEN"
    exit(1)
end

message = ARGV[0]
repo = ENV["GITHUB_REPOSITORY"]

json = File.read(ENV.fetch("GITHUB_EVENT_PATH"))
event = JSON.parse(json)
pr = event["number"]

github = Octokit::Client.new(access_token: ENV["GITHUB_TOKEN"])
comments = github.issue_comments(repo, pr)
comment = comments.find { |c| c["body"].start_with?("Your preview environment") }

if comment
    puts "Message already exists in the PR. Updating"
    github.update_comment(repo, comment["id"], message)
    exit(0)
end

github.add_comment(repo, pr, message)