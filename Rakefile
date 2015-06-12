require './util/rake'

GOPATH=[
	'build/go',
	'go'
]

ENV['GOPATH'] = GOPATH.map { |p|
	"#{Dir.pwd}/#{p}"
}.join(':')

godeps = [
	'github.com/golang/protobuf/...',
	'github.com/jordanorelli/moon'
].map { |pkg|
	path = pkg.gsub(/\.\.\.$/, '')
	dest = "build/go/src/#{path}"
	file dest do
		sh 'go', 'get', pkg
	end
	dest
}

task :default => [
	'build/bin/server'
]

task :edit => godeps do
	sh 'subl', '.'
end

task 'build/bin/server' => godeps + FileList['go/src/pypibot/server/**'] do |t|
	sh 'go', 'build', '-o', t.name, 'pypibot/server'
end

task :nuke do
	rm_rf 'build'
end