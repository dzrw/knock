# encoding: utf-8

require 'pty'

# Script configuration
TARGET = "prod"
SCRIPT_DEBUG_MODE = true

# Things that probably don't need to change...
KNOCK_URL = "mongodb://mongo0002:27017"
KNOCK_OUT_DIR = "/home/deploy/measurements/knock/mongo0002"

KNOCK_PLANS = [

  # GROUP 1
  #
  # "How many concurrent atomic operations can we do during a
  #  a large historics batch containing mostly 1 stream ident?
  #
  # Increment 1 of 10 random counters in a single shared document
  # in order to observe the performance characteristics of atomic
  # updates on various hosts at successive loads.

  # counters, unacknowledged writes
  { mode: 'counters',
    writeconcern: 'none',
    size: 7,
    time: 120,
    moreProps: "-p mongodb.database:knock_counters_unsafe",
  },

  # counters, w=1
  { mode: 'counters',
    writeconcern: 'w=1',
    size: 7,
    time: 120,
    moreProps: "-p mongodb.database:knock_counters_w1",
  },

  # GROUP 2
  #
  # "How fast can we write to the hourly_stats and daily_stats
  #  collections?"
  #
  # Write documents similar to those in hourly_stats and daily_stats
  # in order to observe the performance differences between the
  # various Write Concern settings on various hosts at successive
  # loads.
  #

  # writes, unacknowledged writes, 512 bytes
  { mode: 'writes',
    writeconcern: 'none',
    size: 7,
    time: 60,
    moreProps: "-p mongodb.doc_length:512 -p mongodb.database:knock_writes_unsafe_512b",
    suffix: "-512b"
  },

  # writes, w=0, 512 bytes
  { mode: 'writes',
    writeconcern: 'w=0',
    size: 7,
    time: 60,
    moreProps: "-p mongodb.doc_length:512 -p mongodb.database:knock_writes_w0_512b",
    suffix: "-512b"
  },

  # writes, w=1, 512 bytes
  { mode: 'writes',
    writeconcern: 'w=1',
    size: 7,
    time: 60,
    moreProps: "-p mongodb.doc_length:512 -p mongodb.database:knock_writes_w1_512b",
    suffix: "-512b"
  },

  # GROUP 3
  #
  # "How fast can we write to the exports collection?
  #
  # Write documents similar to those in exports in order to observe
  # the performance differences between the various Write Concern
  # settings on various hosts at successive loads with large
  # documents.
  #

  # writes, unacknowledged writes, 14Kb
  { mode: 'writes',
    writeconcern: 'none',
    size: 5,
    time: 60,
    moreProps: "-p mongodb.doc_length:14336 -p mongodb.database:knock_writes_unsafe_14Kb",
    suffix: "-14Kb"
  },

  # writes, w=0, 14Kb
  { mode: 'writes',
    writeconcern: 'w=0',
    size: 5,
    time: 60,
    moreProps: "-p mongodb.doc_length:14336 -p mongodb.database:knock_writes_w0_14Kb",
    suffix: "-14Kb"
  },

  # writes, w=1, 14Kb
  { mode: 'writes',
    writeconcern: 'w=1',
    size: 5,
    time: 60,
    moreProps: "-p mongodb.doc_length:14336 -p mongodb.database:knock_writes_w1_14Kb",
    suffix: "-14Kb"
  },
]

class KnockWrapper

  attr_reader :plan

  def initialize(plan)
    @plan = plan
  end

  def run
    plan[:size].times do |i|
      command = make_command(plan, 2 ** i)
      res = spawn(command)
      break unless res
    end
  end

  private

  def makepath(mode, load, write_concern, suffix, ext)
    write_concern = write_concern.tr('=', '')
    "#{KNOCK_OUT_DIR}/#{mode}.c#{load}-#{write_concern}#{suffix}.#{ext}"
  end

  def make_command(plan, load)
    opts = [
      "-c #{load}",
      "-d #{plan[:time]}",
      "-v",
      "-p mongodb.url:#{KNOCK_URL}",
      "-p mongodb.run:#{plan[:mode]}",
      "-p mongodb.writeConcern:#{plan[:writeconcern]}",
      "#{plan[:moreProps]|| ""}"
    ]

    args = [
      plan[:mode],
      load,
      plan[:writeconcern],
      plan[:suffix] || "",
      'tsv'
    ]

    { line: "knock #{opts.join(' ')}",
      path: makepath(*args) }
  end

  def spawn(command, io=$stdout)
    (io.puts "#{command[:line]} > #{command[:path]}\n\n"; io.flush) if io
    return true if SCRIPT_DEBUG_MODE

    begin
      PTY.spawn(command[:line]) do |r, stdin, pid|
        done = false

        begin
          File.open(command[:path], 'w') do |fobj|
            r.each do |line|
              fobj.puts(line)
              (io.puts(line); io.flush) if io && !done

              # I don't want to see the histogram in my console.
              if line.start_with?("Time's up!")
                done = true
                (io.puts("\n"); io.flush) if io
              end
            end
          end
        rescue Errno::EIO
          #puts 'Errno:EIO error, but this probably just means ' +
          #      'that the process has finished giving output'
        end
      end
    rescue PTY::ChildExited
      (io.puts 'The child process exited!'; io.flush) if io
      return false
    end

    return true
  end
end

KNOCK_PLANS.each do |plan|
  KnockWrapper.new(plan).run
end

