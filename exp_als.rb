require "fileutils"

projects = Dir["testdata/**/*.als"]
projects.each do |project|
  bn = File.basename(project, ".als")
  FileUtils.cp(project, "misc/#{bn}.gz")
  system %[gunzip -f misc/#{bn}.gz]
  FileUtils.mv("misc/#{bn}", "misc/#{bn}.xml")
end
